package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the main database manager for Athena
type DB struct {
	main *sql.DB
	mu   sync.RWMutex

	dataDir string
}

// New creates a new DB instance and initializes the schema
func New(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "athena.sqlite")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	d := &DB{
		main:    db,
		dataDir: dataDir,
	}

	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return d, nil
}

// Close closes the main database connection
func (d *DB) Close() error {
	return d.main.Close()
}

// DB returns the underlying *sql.DB for the main database
func (d *DB) DB() *sql.DB {
	return d.main
}

// DataDir returns the data directory path
func (d *DB) DataDir() string {
	return d.dataDir
}

// migrate runs all database migrations
func (d *DB) migrate() error {
	migrations := []string{
		migration001Init,
		migration002Auxiliary,
	}

	for i, m := range migrations {
		if _, err := d.main.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}

	return nil
}

const migration001Init = `
-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    original_requirement TEXT,
    requirement_summary TEXT,
    status      TEXT DEFAULT 'active',
    priority    INTEGER DEFAULT 5,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Project goals table
CREATE TABLE IF NOT EXISTS project_goals (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    content     TEXT NOT NULL,
    status      TEXT DEFAULT 'pending',
    assigned_to TEXT REFERENCES agents(id),
    parent_goal TEXT REFERENCES project_goals(id),
    certainty   TEXT DEFAULT 'certain',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Project facts table
CREATE TABLE IF NOT EXISTS project_facts (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    content     TEXT NOT NULL,
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture', 'pending_verification')),
    author      TEXT,
    evidence    TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_project_facts_project ON project_facts(project_id);
CREATE INDEX IF NOT EXISTS idx_project_facts_certainty ON project_facts(certainty);

-- Project discoveries table
CREATE TABLE IF NOT EXISTS project_discoveries (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture')),
    discovered_by TEXT REFERENCES agents(id),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Agents registry
CREATE TABLE IF NOT EXISTS agents (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    role        TEXT NOT NULL,
    status      TEXT DEFAULT 'idle',
    tools       TEXT,
    mcp_servers TEXT,
    model       TEXT DEFAULT 'default',
    state       TEXT,
    created_by  TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_agents_role ON agents(role);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);

-- Project members
CREATE TABLE IF NOT EXISTS project_members (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    agent_id    TEXT NOT NULL REFERENCES agents(id),
    role        TEXT NOT NULL,
    joined_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, agent_id)
);
CREATE INDEX IF NOT EXISTS idx_project_members_project ON project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_agent ON project_members(agent_id);

-- Agent tasks
CREATE TABLE IF NOT EXISTS agent_tasks (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    agent_id    TEXT NOT NULL REFERENCES agents(id),
    title       TEXT NOT NULL,
    description TEXT,
    status      TEXT DEFAULT 'pending',
    priority    INTEGER DEFAULT 5,
    result      TEXT,
    review_status TEXT,
    reviewed_by TEXT REFERENCES agents(id),
    depends_on  TEXT,
    idempotency_key TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_agent_tasks_project ON agent_tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_agent_tasks_agent ON agent_tasks(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_tasks_status ON agent_tasks(status);

-- Agent contexts
CREATE TABLE IF NOT EXISTS agent_contexts (
    id          TEXT PRIMARY KEY,
    agent_id    TEXT NOT NULL REFERENCES agents(id),
    context_type TEXT NOT NULL,
    content     TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Audit log
CREATE TABLE IF NOT EXISTS audit_log (
    id          TEXT PRIMARY KEY,
    agent_id    TEXT NOT NULL,
    action      TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id   TEXT NOT NULL,
    details     TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_audit_log_agent ON audit_log(agent_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_target ON audit_log(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_time ON audit_log(created_at);
`

const migration002Auxiliary = `
-- Company config table (replaces athena.yaml for mutable settings)
CREATE TABLE IF NOT EXISTS company_config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Insert default company limits
INSERT OR IGNORE INTO company_config (key, value) VALUES ('max_agents', '100');
INSERT OR IGNORE INTO company_config (key, value) VALUES ('max_memory_mb', '16384');
INSERT OR IGNORE INTO company_config (key, value) VALUES ('cost_budget', '0');
INSERT OR IGNORE INTO company_config (key, value) VALUES ('cost_spent', '0');
`
