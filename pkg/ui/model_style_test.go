package ui

import (
	"strings"
	"testing"

	"github.com/vanderheijden86/beadwork/pkg/model"
)

func TestModelStylePlaceholder(t *testing.T) {
	// Placeholder to keep file non-empty after reverting footer-style experiments.
}

func TestTreeFooterShowsManualRefreshShortcut(t *testing.T) {
	m := NewModel([]model.Issue{{ID: "test-1", Title: "Test", Status: model.StatusOpen}}, "")
	m.width = 200

	footer := stripANSI(m.renderFooter())
	if !strings.Contains(footer, "^R:refresh") {
		t.Fatalf("expected tree footer to expose manual refresh shortcut, got %q", footer)
	}
}
