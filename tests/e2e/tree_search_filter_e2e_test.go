// tree_search_filter_e2e_test.go - End-to-end proof that free-text search keeps
// an active label filter instead of dropping it (bd-oe1y).
package main_test

import (
	"strings"
	"testing"
	"time"
)

// makeLabelSearchFixture builds a fixture where the "bug" label is the most used
// one, so it is reachable as label number 1 in the picker (labels are numbered by
// count descending). Only one "bug" issue also matches the query "alpha".
func makeLabelSearchFixture(t *testing.T) []treeFixtureIssue {
	t.Helper()
	now := time.Now()
	return []treeFixtureIssue{
		{ID: "bug-alpha", Title: "Alpha bug work", Status: "open", Priority: 1, IssueType: "bug",
			CreatedAt: now.Format(time.RFC3339), Labels: []string{"bug"}},
		{ID: "bug-beta", Title: "Beta bug work", Status: "open", Priority: 1, IssueType: "bug",
			CreatedAt: now.Add(time.Second).Format(time.RFC3339), Labels: []string{"bug"}},
		{ID: "bug-gamma", Title: "Gamma bug work", Status: "open", Priority: 1, IssueType: "bug",
			CreatedAt: now.Add(2 * time.Second).Format(time.RFC3339), Labels: []string{"bug"}},
		{ID: "feat-alpha", Title: "Alpha feature work", Status: "open", Priority: 2, IssueType: "feature",
			CreatedAt: now.Add(3 * time.Second).Format(time.RFC3339), Labels: []string{"feature"}},
		{ID: "feat-beta", Title: "Beta feature work", Status: "open", Priority: 2, IssueType: "feature",
			CreatedAt: now.Add(4 * time.Second).Format(time.RFC3339), Labels: []string{"feature"}},
	}
}

func TestSearchFieldHiddenUntilSlashE2E(t *testing.T) {
	tempDir := t.TempDir()
	writeTreeFixture(t, tempDir, makeLabelSearchFixture(t))

	idleOut, err := runTreeTUI(t, tempDir, 1200, nil)
	if err != nil {
		t.Fatalf("idle TUI run failed: %v\noutput:\n%s", err, idleOut)
	}
	if strings.Contains(string(idleOut), "/█") {
		t.Fatalf("idle TUI rendered the search field\noutput:\n%s", idleOut)
	}

	editingOut, err := runTreeTUI(t, tempDir, 1500, []keyStep{kd("/", 150*time.Millisecond)})
	if err != nil {
		t.Fatalf("editing TUI run failed: %v\noutput:\n%s", err, editingOut)
	}
	if !strings.Contains(string(editingOut), "/█") {
		t.Fatalf("slash did not reveal the search field\noutput:\n%s", editingOut)
	}
}

func TestSearchTabCompletesPlainLabelE2E(t *testing.T) {
	tempDir := t.TempDir()
	writeTreeFixture(t, tempDir, []treeFixtureIssue{
		{ID: "dispatch-1", Title: "Dispatch", Status: "open", Priority: 1, IssueType: "task", CreatedAt: time.Now().Format(time.RFC3339), Labels: []string{"lane-attempt=1"}},
	})

	out, err := runTreeTUI(t, tempDir, 2200, []keyStep{
		kd("/", 150*time.Millisecond),
		kd("l", 80*time.Millisecond),
		kd("a", 80*time.Millisecond),
		kd("n", 80*time.Millisecond),
		kd("\t", 100*time.Millisecond),
	})
	if err != nil {
		t.Fatalf("completion TUI run failed: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(string(out), "/ lane-attempt=1█") {
		t.Fatalf("Tab did not complete the plain label\noutput:\n%s", out)
	}
}

func TestEmptyLabelFacetKeepsResultsE2E(t *testing.T) {
	tempDir := t.TempDir()
	writeTreeFixture(t, tempDir, makeLabelSearchFixture(t))

	out, err := runTreeTUI(t, tempDir, 2200, []keyStep{
		kd("/", 150*time.Millisecond),
		kd("l", 60*time.Millisecond),
		kd("a", 60*time.Millisecond),
		kd("b", 60*time.Millisecond),
		kd("e", 60*time.Millisecond),
		kd("l", 60*time.Millisecond),
		kd(":", 60*time.Millisecond),
		kd("\r", 100*time.Millisecond),
		kd("K", 100*time.Millisecond),
	})
	if err != nil {
		t.Fatalf("empty-facet TUI run failed: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(string(out), "Close issue?") {
		t.Fatalf("empty label facet removed every selectable issue\noutput:\n%s", out)
	}
}

// TestTreeSearchKeepsLabelFilterE2E drives the real TUI through a composed
// label filter and free-text query.
func TestTreeSearchKeepsLabelFilterE2E(t *testing.T) {
	tempDir := t.TempDir()
	writeTreeFixture(t, tempDir, makeLabelSearchFixture(t))

	out, err := runTreeTUI(t, tempDir, 4000, []keyStep{
		kd("L", 150*time.Millisecond), // switch picker to label mode
		kd("1", 150*time.Millisecond), // filter to the "bug" label
		kd("/", 150*time.Millisecond), // enter tree search
		kd("a", 80*time.Millisecond),
		kd("l", 80*time.Millisecond),
		kd("p", 80*time.Millisecond),
		kd("h", 80*time.Millisecond),
		kd("a", 80*time.Millisecond),
	})
	if err != nil {
		t.Fatalf("TUI run failed: %v\noutput:\n%s", err, out)
	}

	s := string(out)
	if !strings.Contains(s, "/ alpha█") {
		t.Errorf("shared search field never showed the full query %q\noutput:\n%s", "alpha", s)
	}
	if !strings.Contains(s, "╭") || !strings.Contains(s, "╰") {
		t.Errorf("expected a bordered search field\noutput:\n%s", s)
	}
	if strings.Contains(s, "WHERE / FILTER") || strings.Contains(s, "ORDER BY") {
		t.Errorf("search field contains query-plan clutter\noutput:\n%s", s)
	}
}

// TestGlobalFuzzyLabelSearchE2E verifies that plain text can find an issue by
// an abbreviated label even when the title and ID do not contain the query.
func TestGlobalFuzzyLabelSearchE2E(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Now()
	writeTreeFixture(t, tempDir, []treeFixtureIssue{
		{ID: "dispatch-1", Title: "Dispatch first lane", Status: "open", Priority: 1, IssueType: "task",
			CreatedAt: now.Format(time.RFC3339), Labels: []string{"lane-attempt=1"}},
		{ID: "queue-1", Title: "Queue maintenance", Status: "open", Priority: 2, IssueType: "task",
			CreatedAt: now.Add(time.Second).Format(time.RFC3339), Labels: []string{"queue"}},
	})

	out, err := runTreeTUI(t, tempDir, 3000, []keyStep{
		kd("/", 150*time.Millisecond),
		kd("l", 80*time.Millisecond),
		kd("n", 80*time.Millisecond),
		kd("a", 80*time.Millisecond),
		kd("1", 80*time.Millisecond),
	})
	if err != nil {
		t.Fatalf("TUI run failed: %v\noutput:\n%s", err, out)
	}

	s := string(out)
	if !strings.Contains(s, "/ lna1█") {
		t.Errorf("search field never showed fuzzy label query\noutput:\n%s", s)
	}
	if strings.Contains(s, "No issues to display.") {
		t.Errorf("fuzzy label query produced an empty result set\noutput:\n%s", s)
	}
}
