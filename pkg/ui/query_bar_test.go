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
	for _, want := range []string{"WHERE / FILTER", "id:bv-3", "1/4", "ORDER BY"} {
		if !strings.Contains(bar, want) {
			t.Errorf("query bar %q does not contain %q", bar, want)
		}
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

func TestQueryBarShowsLegacyFiltersAsChips(t *testing.T) {
	m := newSearchFilterModel(t, "ui")
	m.currentFilter = "open"
	m.assigneeFilter = "ann"
	m.setQueryText("id:bv")

	bar := stripANSI(m.renderUnifiedTitleBar(160))
	for _, want := range []string{"[status:open]", "[label:ui]", "[assignee:ann]"} {
		if !strings.Contains(bar, want) {
			t.Errorf("query bar %q does not contain filter chip %q", bar, want)
		}
	}
}

func TestTreeQuickFilterFeedsSharedBar(t *testing.T) {
	m := newSearchFilterModel(t, "")
	m = typeKeys(m, "o")

	if got := m.currentFilter; got != "open" {
		t.Fatalf("model filter = %q, want open", got)
	}
	bar := stripANSI(m.renderUnifiedTitleBar(120))
	if !strings.Contains(bar, "[status:open]") {
		t.Fatalf("query bar %q does not show the tree's status filter", bar)
	}
}
