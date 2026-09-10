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

// TestTreeSearchKeepsLabelFilterE2E drives the real TUI: apply a label filter,
// then type a free-text query containing "a".
//
// Before the fix the "a" keystroke fired the global "clear all filters"
// shortcut, so the query came out as "lpha" and the label filter was gone.
// The shared query bar shows the full query, label chip, and one live result.
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
	if !strings.Contains(s, "WHERE / FILTER") || !strings.Contains(s, "/ alpha") {
		t.Errorf("shared query bar never showed the full query %q\noutput:\n%s", "alpha", s)
	}
	if !strings.Contains(s, "[label:bug]") || !strings.Contains(s, "1/5") {
		t.Errorf("expected the bar to show the active label and one result out of five\noutput:\n%s", s)
	}
	if !strings.Contains(s, "ORDER BY") {
		t.Errorf("expected the query bar to show the active ordering\noutput:\n%s", s)
	}
}
