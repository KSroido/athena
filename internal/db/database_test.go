package db

import (
	"os"
	"testing"
)

func TestNewDB(t *testing.T) {
	dir := t.TempDir()
	d, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer d.Close()

	// Verify tables exist
	var count int
	err = d.DB().QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	if err != nil {
		t.Fatalf("query tables: %v", err)
	}
	if count < 5 {
		t.Errorf("expected at least 5 tables, got %d", count)
	}
}

func TestCreateProject(t *testing.T) {
	dir := t.TempDir()
	d, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer d.Close()

	_, err = d.DB().Exec(`
		INSERT INTO projects (id, name, original_requirement, status)
		VALUES ('test-001', 'Test Project', 'Build a test app', 'active')
	`)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}

	var name string
	err = d.DB().QueryRow("SELECT name FROM projects WHERE id = 'test-001'").Scan(&name)
	if err != nil {
		t.Fatalf("query project: %v", err)
	}
	if name != "Test Project" {
		t.Errorf("expected 'Test Project', got '%s'", name)
	}
}

func TestCreateAgent(t *testing.T) {
	dir := t.TempDir()
	d, err := New(dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer d.Close()

	_, err = d.DB().Exec(`
		INSERT INTO agents (id, name, role, status, model)
		VALUES ('test-pm-1', 'PM Agent', 'pm', 'idle', 'default')
	`)
	if err != nil {
		t.Fatalf("insert agent: %v", err)
	}

	var role string
	err = d.DB().QueryRow("SELECT role FROM agents WHERE id = 'test-pm-1'").Scan(&role)
	if err != nil {
		t.Fatalf("query agent: %v", err)
	}
	if role != "pm" {
		t.Errorf("expected 'pm', got '%s'", role)
	}
}

func TestMain(t *testing.T) {
	// Verify the test runs with a real temp dir
	dir := t.TempDir()
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("temp dir not accessible: %v", err)
	}
}
