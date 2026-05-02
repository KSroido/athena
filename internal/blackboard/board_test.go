package blackboard

import (
	"testing"

	"github.com/ksroido/athena/internal/db"
)

func TestOpenBoard(t *testing.T) {
	dir := t.TempDir()
	board, err := OpenBoard(dir, "test-proj")
	if err != nil {
		t.Fatalf("OpenBoard() error: %v", err)
	}
	defer board.Close()

	if board.ProjectID() != "test-proj" {
		t.Errorf("expected project ID 'test-proj', got '%s'", board.ProjectID())
	}
}

func TestWriteAndReadEntry(t *testing.T) {
	dir := t.TempDir()
	board, err := OpenBoard(dir, "test-proj")
	if err != nil {
		t.Fatalf("OpenBoard() error: %v", err)
	}
	defer board.Close()

	entry := &db.BlackboardEntry{
		ID:        "entry-001",
		ProjectID: "test-proj",
		Category:  CategoryFact,
		Content:   "Project uses Go 1.24",
		Certainty: CertaintyCertain,
		Author:    "test-dev-1",
	}

	err = board.WriteEntrySync(entry)
	if err != nil {
		t.Fatalf("WriteEntrySync() error: %v", err)
	}

	entries, err := board.ReadEntries(CategoryFact, 10, 0)
	if err != nil {
		t.Fatalf("ReadEntries() error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Content != "Project uses Go 1.24" {
		t.Errorf("expected 'Project uses Go 1.24', got '%s'", entries[0].Content)
	}
	if entries[0].Certainty != CertaintyCertain {
		t.Errorf("expected certainty 'certain', got '%s'", entries[0].Certainty)
	}
}

func TestSearchEntries(t *testing.T) {
	dir := t.TempDir()
	board, err := OpenBoard(dir, "test-proj")
	if err != nil {
		t.Fatalf("OpenBoard() error: %v", err)
	}
	defer board.Close()

	// Write two entries
	board.WriteEntrySync(&db.BlackboardEntry{
		ID:        "e1",
		ProjectID: "test-proj",
		Category:  CategoryFact,
		Content:   "Go version is 1.24",
		Certainty: CertaintyCertain,
		Author:    "dev-1",
	})
	board.WriteEntrySync(&db.BlackboardEntry{
		ID:        "e2",
		ProjectID: "test-proj",
		Category:  CategoryDiscovery,
		Content:   "Found SQLite WAL mode issue",
		Certainty: CertaintyConjecture,
		Author:    "dev-1",
	})

	// Search for "Go"
	results, err := board.Search("Go", 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(results) < 1 {
		t.Errorf("expected at least 1 result for 'Go', got %d", len(results))
	}
}

func TestAccessControl(t *testing.T) {
	// Developer can read facts
	if !CanRead("developer", Level1Facts) {
		t.Error("developer should be able to read facts")
	}

	// Developer can write discoveries
	if !CanWrite("developer", Level4Discovery) {
		t.Error("developer should be able to write discoveries")
	}

	// HR cannot write facts
	if CanWrite("hr", Level1Facts) {
		t.Error("HR should not be able to write facts")
	}

	// PM can read and write most levels
	if !CanWrite("pm", Level0Meta) {
		t.Error("PM should be able to write meta")
	}

	// Reviewer can write facts but not progress
	if !CanWrite("reviewer", Level1Facts) {
		t.Error("reviewer should be able to write facts")
	}
	if CanWrite("reviewer", Level3Progress) {
		t.Error("reviewer should not be able to write progress")
	}
}

func TestCategoryToLevel(t *testing.T) {
	tests := map[string]int{
		CategoryGoal:       Level0Meta,
		CategoryFact:       Level1Facts,
		CategoryProgress:   Level3Progress,
		CategoryDiscovery:  Level4Discovery,
		CategoryResolution: Level5Resolution,
		CategoryAuxiliary:  Level4_5Aux,
	}

	for cat, expected := range tests {
		got := CategoryToLevel(cat)
		if got != expected {
			t.Errorf("CategoryToLevel(%s) = %d, want %d", cat, got, expected)
		}
	}
}
