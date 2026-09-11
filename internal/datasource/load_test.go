package datasource

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDiscoveryOptionsSkipPreflightValidation(t *testing.T) {
	opts := loadDiscoveryOptions("/repo/.beads", "/repo")

	if opts.ValidateAfterDiscovery {
		t.Fatal("load discovery must not validate a source before LoadFromSource opens and queries it")
	}
	if opts.IncludeInvalid {
		t.Fatal("load discovery should not request invalid-source metadata")
	}
}

func TestLoadSmartLoadsAnUnvalidatedCandidate(t *testing.T) {
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	issueData := []byte("{\"id\":\"test-1\",\"title\":\"Fast path\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}\n")
	if err := os.WriteFile(filepath.Join(beadsDir, "issues.jsonl"), issueData, 0o644); err != nil {
		t.Fatal(err)
	}

	issues, err := loadSmart(beadsDir, "")
	if err != nil {
		t.Fatalf("loadSmart failed: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != "test-1" {
		t.Fatalf("unexpected issues: %#v", issues)
	}
}

func TestLoadSmartIgnoresNewerAuxiliaryJSONL(t *testing.T) {
	repoDir := t.TempDir()
	beadsDir := filepath.Join(repoDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	issuesPath := filepath.Join(beadsDir, "issues.jsonl")
	issueData := []byte("{\"id\":\"test-1\",\"title\":\"Canonical issue\",\"status\":\"open\",\"priority\":2,\"issue_type\":\"task\"}\n")
	if err := os.WriteFile(issuesPath, issueData, 0o644); err != nil {
		t.Fatal(err)
	}

	interactionsPath := filepath.Join(beadsDir, "interactions.jsonl")
	interactionData := []byte("{\"id\":\"int-1\",\"kind\":\"field_change\",\"issue_id\":\"test-1\"}\n")
	if err := os.WriteFile(interactionsPath, interactionData, 0o644); err != nil {
		t.Fatal(err)
	}

	older := time.Unix(1_700_000_000, 0)
	newer := older.Add(time.Hour)
	if err := os.Chtimes(issuesPath, older, older); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(interactionsPath, newer, newer); err != nil {
		t.Fatal(err)
	}

	issues, err := loadSmart(beadsDir, repoDir)
	if err != nil {
		t.Fatalf("loadSmart failed: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != "test-1" {
		t.Fatalf("unexpected issues: %#v", issues)
	}
}
