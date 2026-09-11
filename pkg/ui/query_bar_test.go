package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSlashStartsSharedQueryEditingInEveryView(t *testing.T) {
	tests := []struct {
		name  string
		focus focus
		setup func(*Model)
	}{
		{name: "tree", focus: focusTree, setup: func(m *Model) { m.treeViewActive = true }},
		{name: "list", focus: focusList, setup: func(m *Model) { m.treeViewActive = false }},
		{name: "board", focus: focusBoard, setup: func(m *Model) { m.isBoardView = true }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSearchFilterModel(t, "")
			m.focused = tt.focus
			tt.setup(&m)

			m = typeKeys(m, "/")

			if got := m.queryState.Mode(); got != QueryEditing {
				t.Fatalf("query mode = %v, want editing", got)
			}
			if m.tree.IsSearchMode() {
				t.Error("tree-local search mode should not own the shared query")
			}
			if m.board.IsSearchMode() {
				t.Error("board-local search mode should not own the shared query")
			}
		})
	}
}

func TestSharedQueryFiltersByIDAndPersistsAfterAccept(t *testing.T) {
	m := newSearchFilterModel(t, "")
	m = typeKeys(m, "/", "i", "d", ":", "b", "v", "-", "3")

	if got := m.queryState.Text(); got != "id:bv-3" {
		t.Fatalf("query text = %q, want id:bv-3", got)
	}
	if got := len(m.list.Items()); got != 1 {
		t.Fatalf("filtered list count = %d, want 1", got)
	}
	if got := m.tree.NodeCount(); got != 1 {
		t.Fatalf("filtered tree count = %d, want 1", got)
	}
	if got := m.board.SearchQuery(); got != "id:bv-3" {
		t.Fatalf("board query = %q, want id:bv-3", got)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if got := m.queryState.Mode(); got != QueryIdle {
		t.Fatalf("query mode = %v, want idle after enter", got)
	}
	if got := m.queryState.Text(); got != "id:bv-3" {
		t.Fatalf("accepted query = %q, want it to persist", got)
	}

	bar := stripANSI(m.renderUnifiedTitleBar(120))
	if !strings.Contains(bar, "/ id:bv-3") {
		t.Errorf("query bar %q does not contain the accepted search", bar)
	}
}

func TestEscapeClearsAcceptedSharedQuery(t *testing.T) {
	m := newSearchFilterModel(t, "")
	m = typeKeys(m, "/", "b", "e", "t", "a")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if !m.queryState.Empty() || m.queryState.Text() != "" {
		t.Fatalf("query was not cleared: %q", m.queryState.Text())
	}
	if got := len(m.list.Items()); got != 4 {
		t.Fatalf("list count after clearing = %d, want 4", got)
	}
	if got := m.tree.NodeCount(); got != 4 {
		t.Fatalf("tree count after clearing = %d, want 4", got)
	}
}

func TestQueryTabCompletesFieldsAndValues(t *testing.T) {
	m := newSearchFilterModel(t, "")
	m = typeKeys(m, "/", "s", "t", "a")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if got := m.queryState.Text(); got != "status:" {
		t.Fatalf("field completion = %q, want status:", got)
	}

	m = typeKeys(m, "o")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if got := m.queryState.Text(); got != "status:open" {
		t.Fatalf("value completion = %q, want status:open", got)
	}
}

func TestQueryBarKeepsSearchReadableWithLegacyFilters(t *testing.T) {
	m := newSearchFilterModel(t, "ui")
	m.currentFilter = "open"
	m.assigneeFilter = "ann"
	m.setQueryText("id:bv")

	bar := stripANSI(m.renderUnifiedTitleBar(160))
	if !strings.Contains(bar, "/ id:bv") {
		t.Errorf("query bar %q does not contain the active search", bar)
	}
	for _, hiddenFilter := range []string{"[status:open]", "[label:ui]", "[assignee:ann]"} {
		if strings.Contains(bar, hiddenFilter) {
			t.Errorf("query bar %q contains legacy filter chip %q", bar, hiddenFilter)
		}
	}
}

func TestQueryBarUsesCleanBorderedSearchField(t *testing.T) {
	m := newSearchFilterModel(t, "ui")
	m.currentFilter = "open"
	m.assigneeFilter = "ann"
	m.setQueryText("lane-a")
	m.queryState.StartEditing()

	bar := stripANSI(m.renderUnifiedTitleBar(80))
	lines := strings.Split(bar, "\n")
	if len(lines) != 3 {
		t.Fatalf("search field height = %d, want 3 lines: %q", len(lines), bar)
	}
	if !strings.HasPrefix(lines[0], "╭") || !strings.HasSuffix(lines[0], "╮") {
		t.Errorf("search field top border is incomplete: %q", lines[0])
	}
	if !strings.HasPrefix(lines[2], "╰") || !strings.HasSuffix(lines[2], "╯") {
		t.Errorf("search field bottom border is incomplete: %q", lines[2])
	}
	if !strings.Contains(lines[1], "/ lane-a█") {
		t.Errorf("search field does not contain query and cursor: %q", lines[1])
	}
	for _, clutter := range []string{"WHERE / FILTER", "ORDER BY", "[status:", "[label:", "[assignee:", "[project:", "⇥"} {
		if strings.Contains(bar, clutter) {
			t.Errorf("search field contains %q clutter: %q", clutter, bar)
		}
	}
}

func TestTreeQuickFilterRemainsActiveWithCleanSearchBar(t *testing.T) {
	m := newSearchFilterModel(t, "")
	m = typeKeys(m, "o")

	if got := m.currentFilter; got != "open" {
		t.Fatalf("model filter = %q, want open", got)
	}
	bar := stripANSI(m.renderUnifiedTitleBar(120))
	if strings.Contains(bar, "[status:open]") {
		t.Fatalf("query bar %q includes status-filter clutter", bar)
	}
}
