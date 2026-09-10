package ui

import (
	"testing"

	"github.com/vanderheijden86/beadwork/pkg/model"
)

func TestIssueQueryPlainTextMatchesIDAndTitle(t *testing.T) {
	issue := model.Issue{ID: "bd-sxgr", Title: "Live refresh support"}

	for _, raw := range []string{"sxgr", "live refresh"} {
		if query := ParseIssueQuery(raw); !query.Matches(issue) {
			t.Errorf("query %q did not match issue ID or title", raw)
		}
	}
	if query := ParseIssueQuery("unrelated"); query.Matches(issue) {
		t.Error("unrelated plain query matched issue")
	}
}

func TestIssueQueryUsesORWithinFacetAndANDBetweenFacets(t *testing.T) {
	query := ParseIssueQuery("status:open status:blocked type:epic")

	openEpic := model.Issue{Status: model.StatusOpen, IssueType: model.TypeEpic}
	blockedEpic := model.Issue{Status: model.StatusBlocked, IssueType: model.TypeEpic}
	openTask := model.Issue{Status: model.StatusOpen, IssueType: model.TypeTask}

	if !query.Matches(openEpic) || !query.Matches(blockedEpic) {
		t.Error("same-facet status values should use OR semantics")
	}
	if query.Matches(openTask) {
		t.Error("different facets should use AND semantics")
	}
}

func TestIssueQueryMatchesSupportedFields(t *testing.T) {
	issue := model.Issue{
		ID:         "bd-7rt1.3",
		Title:      "Add filter palette",
		Status:     model.StatusOpen,
		Priority:   2,
		IssueType:  model.TypeFeature,
		Labels:     []string{"ui", "search"},
		Assignee:   "andre",
		SourceRepo: "b9s",
	}

	for _, raw := range []string{
		"id:7rt1",
		"title:palette",
		"status:open",
		"priority:2",
		"type:feature",
		"label:search",
		"assignee:andre",
		"project:b9s",
		"!status:closed",
	} {
		if query := ParseIssueQuery(raw); !query.Matches(issue) {
			t.Errorf("supported query %q did not match issue", raw)
		}
	}
}

func TestQueryStateHasFiniteEditingLifecycle(t *testing.T) {
	var state QueryState
	if state.Mode() != QueryIdle {
		t.Fatalf("initial mode = %v, want idle", state.Mode())
	}

	state.StartEditing()
	state.SetText("id:sxgr")
	if state.Mode() != QueryEditing {
		t.Fatalf("editing mode = %v, want editing", state.Mode())
	}
	state.Accept()
	if state.Mode() != QueryIdle || state.Text() != "id:sxgr" {
		t.Fatalf("accepted state = (%v, %q), want idle with retained query", state.Mode(), state.Text())
	}
	state.Clear()
	if !state.Empty() || state.Mode() != QueryIdle {
		t.Fatalf("cleared state = (%v, %q), want empty idle", state.Mode(), state.Text())
	}
}

func TestModelQueryFiltersTreeListAndBoard(t *testing.T) {
	issues := []model.Issue{
		{ID: "bd-target", Title: "Target epic", Status: model.StatusOpen, IssueType: model.TypeEpic},
		{ID: "bd-other", Title: "Other task", Status: model.StatusOpen, IssueType: model.TypeTask},
	}
	m := NewModel(issues, "")
	m.tree.Build(issues)

	m.setQueryText("id:target")

	if got := len(m.list.Items()); got != 1 {
		t.Errorf("list items = %d, want 1", got)
	}
	if got := len(m.board.allIssues); got != 1 {
		t.Errorf("board issues = %d, want 1", got)
	}
	if got := m.tree.NodeCount(); got != 1 {
		t.Errorf("tree nodes = %d, want 1", got)
	}
}
