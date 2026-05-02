package blackboard

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ksroido/athena/internal/db"
)

// Certainty levels for blackboard entries
const (
	CertaintyCertain            = "certain"
	CertaintyConjecture         = "conjecture"
	CertaintyPendingVerification = "pending_verification"
)

// Categories for blackboard entries
const (
	CategoryGoal              = "goal"
	CategoryFact              = "fact"
	CategoryDiscovery         = "discovery"
	CategoryDecision          = "decision"
	CategoryProgress          = "progress"
	CategoryResolution        = "resolution"
	CategoryAuxiliary         = "auxiliary"
	CategoryAcceptanceCrit    = "acceptance_criteria" // PM-defined acceptance criteria for verification
	CategoryVerification      = "verification"         // Verification round results (submit/review/escalation)
)

// Board manages a project-specific blackboard database
type Board struct {
	projectID string
	db        *sql.DB
	mu        sync.RWMutex
	dataDir   string

	// Channel for batched writes
	writeCh   chan *db.BlackboardEntry
	batchSize int
	done      chan struct{}
}

// OpenBoard opens (or creates) a blackboard for the given project
func OpenBoard(dataDir, projectID string) (*Board, error) {
	boardDir := filepath.Join(dataDir, "board")
	if err := os.MkdirAll(boardDir, 0755); err != nil {
		return nil, fmt.Errorf("create board dir: %w", err)
	}

	dbPath := filepath.Join(boardDir, fmt.Sprintf("board_%s.sqlite", projectID))
	d, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open board db: %w", err)
	}

	if _, err := d.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := d.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	b := &Board{
		projectID: projectID,
		db:        d,
		dataDir:   dataDir,
		writeCh:   make(chan *db.BlackboardEntry, 256),
		batchSize: 32,
		done:      make(chan struct{}),
	}

	if err := b.migrate(); err != nil {
		return nil, fmt.Errorf("migrate board: %w", err)
	}

	// Start the batch writer goroutine
	go b.batchWriter()

	return b, nil
}

// Close closes the board database and stops the batch writer
func (b *Board) Close() error {
	close(b.done)
	return b.db.Close()
}

// ProjectID returns the project ID for this board
func (b *Board) ProjectID() string {
	return b.projectID
}

// WriteEntry queues an entry for writing to the blackboard
// The actual write is batched via channel for performance
func (b *Board) WriteEntry(entry *db.BlackboardEntry) {
	entry.ProjectID = b.projectID
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()
	if entry.LastVerifiedAt == nil {
		now := time.Now()
		entry.LastVerifiedAt = &now
	}
	b.writeCh <- entry
}

// WriteEntrySync writes an entry to the blackboard synchronously
func (b *Board) WriteEntrySync(entry *db.BlackboardEntry) error {
	entry.ProjectID = b.projectID
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()
	if entry.LastVerifiedAt == nil {
		now := time.Now()
		entry.LastVerifiedAt = &now
	}
	return b.insertEntry(entry)
}

// ReadEntries reads blackboard entries with optional category filter
func (b *Board) ReadEntries(category string, limit, offset int) ([]*db.BlackboardEntry, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = b.db.Query(
			"SELECT id, project_id, category, content, certainty, author, confidence_score, reasoning, last_verified_at, created_at, updated_at FROM blackboard_entries WHERE category = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?",
			category, limit, offset,
		)
	} else {
		rows, err = b.db.Query(
			"SELECT id, project_id, category, content, certainty, author, confidence_score, reasoning, last_verified_at, created_at, updated_at FROM blackboard_entries ORDER BY updated_at DESC LIMIT ? OFFSET ?",
			limit, offset,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	var entries []*db.BlackboardEntry
	for rows.Next() {
		e := &db.BlackboardEntry{}
		var lastVerified sql.NullTime
		var confidenceScore sql.NullInt64

		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Category, &e.Content, &e.Certainty,
			&e.Author, &confidenceScore, &e.Reasoning, &lastVerified,
			&e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan entry: %w", err)
		}

		if confidenceScore.Valid {
			score := int(confidenceScore.Int64)
			e.ConfidenceScore = &score
		}
		if lastVerified.Valid {
			e.LastVerifiedAt = &lastVerified.Time
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// Search performs full-text search on blackboard entries
func (b *Board) Search(query string, limit int) ([]*db.BlackboardEntry, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	rows, err := b.db.Query(`
		SELECT e.id, e.project_id, e.category, e.content, e.certainty, e.author,
		       e.confidence_score, e.reasoning, e.last_verified_at, e.created_at, e.updated_at
		FROM blackboard_entries e
		JOIN blackboard_entries_fts fts ON e.rowid = fts.rowid
		WHERE blackboard_entries_fts MATCH ?
		ORDER BY fts.rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fts search: %w", err)
	}
	defer rows.Close()

	var entries []*db.BlackboardEntry
	for rows.Next() {
		e := &db.BlackboardEntry{}
		var lastVerified sql.NullTime
		var confidenceScore sql.NullInt64

		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Category, &e.Content, &e.Certainty,
			&e.Author, &confidenceScore, &e.Reasoning, &lastVerified,
			&e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}

		if confidenceScore.Valid {
			score := int(confidenceScore.Int64)
			e.ConfidenceScore = &score
		}
		if lastVerified.Valid {
			e.LastVerifiedAt = &lastVerified.Time
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// DeleteEntry deletes a blackboard entry (only the author can delete)
func (b *Board) DeleteEntry(entryID, authorID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	result, err := b.db.Exec("DELETE FROM blackboard_entries WHERE id = ? AND author = ?", entryID, authorID)
	if err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("entry not found or not owned by %s", authorID)
	}

	return nil
}

// batchWriter reads from the write channel and batch-commits entries to SQLite
func (b *Board) batchWriter() {
	batch := make([]*db.BlackboardEntry, 0, b.batchSize)
	ticker := time.NewTicker(500 * time.Millisecond) // flush every 500ms
	defer ticker.Stop()

	for {
		select {
		case entry := <-b.writeCh:
			batch = append(batch, entry)
			if len(batch) >= b.batchSize {
				b.flushBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				b.flushBatch(batch)
				batch = batch[:0]
			}

		case <-b.done:
			// Flush remaining entries before closing
			if len(batch) > 0 {
				b.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch writes a batch of entries to SQLite in a single transaction
func (b *Board) flushBatch(batch []*db.BlackboardEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	tx, err := b.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO blackboard_entries (id, project_id, category, content, certainty, author, confidence_score, reasoning, last_verified_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, e := range batch {
		var confidenceScore sql.NullInt64
		if e.ConfidenceScore != nil {
			confidenceScore = sql.NullInt64{Int64: int64(*e.ConfidenceScore), Valid: true}
		}
		var lastVerified sql.NullString
		if e.LastVerifiedAt != nil {
			lastVerified = sql.NullString{String: e.LastVerifiedAt.Format(time.RFC3339), Valid: true}
		}

		_, err := stmt.Exec(e.ID, e.ProjectID, e.Category, e.Content, e.Certainty,
			e.Author, confidenceScore, e.Reasoning, lastVerified,
			e.CreatedAt.Format(time.RFC3339), e.UpdatedAt.Format(time.RFC3339))
		if err != nil {
			continue // skip failed entries
		}
	}

	tx.Commit()
}

// insertEntry inserts a single entry synchronously
func (b *Board) insertEntry(entry *db.BlackboardEntry) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var confidenceScore sql.NullInt64
	if entry.ConfidenceScore != nil {
		confidenceScore = sql.NullInt64{Int64: int64(*entry.ConfidenceScore), Valid: true}
	}
	var lastVerified sql.NullString
	if entry.LastVerifiedAt != nil {
		lastVerified = sql.NullString{String: entry.LastVerifiedAt.Format(time.RFC3339), Valid: true}
	}

	_, err := b.db.Exec(`
		INSERT INTO blackboard_entries (id, project_id, category, content, certainty, author, confidence_score, reasoning, last_verified_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		entry.ID, entry.ProjectID, entry.Category, entry.Content, entry.Certainty,
		entry.Author, confidenceScore, entry.Reasoning, lastVerified,
		entry.CreatedAt.Format(time.RFC3339), entry.UpdatedAt.Format(time.RFC3339),
	)

	return err
}

// migrate creates the blackboard schema
func (b *Board) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS blackboard_entries (
		id          TEXT PRIMARY KEY,
		project_id  TEXT NOT NULL,
		category    TEXT NOT NULL,
		content     TEXT NOT NULL,
		certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture', 'pending_verification')),
		author      TEXT,
		confidence_score INTEGER,
		reasoning   TEXT,
		last_verified_at DATETIME,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_bb_category ON blackboard_entries(category);
	CREATE INDEX IF NOT EXISTS idx_bb_certainty ON blackboard_entries(certainty);
	CREATE INDEX IF NOT EXISTS idx_bb_author ON blackboard_entries(author);

	-- FTS5 full-text search
	CREATE VIRTUAL TABLE IF NOT EXISTS blackboard_entries_fts USING fts5(
		content,
		category,
		author,
		content=blackboard_entries,
		content_rowid=rowid,
		tokenize='unicode61'
	);

	-- Triggers for automatic FTS index updates
	CREATE TRIGGER IF NOT EXISTS blackboard_entries_ai AFTER INSERT ON blackboard_entries BEGIN
		INSERT INTO blackboard_entries_fts(rowid, content, category, author)
		VALUES (new.rowid, new.content, new.category, new.author);
	END;

	CREATE TRIGGER IF NOT EXISTS blackboard_entries_ad AFTER DELETE ON blackboard_entries BEGIN
		INSERT INTO blackboard_entries_fts(blackboard_entries_fts, rowid, content, category, author)
		VALUES ('delete', old.rowid, old.content, old.category, old.author);
	END;

	CREATE TRIGGER IF NOT EXISTS blackboard_entries_au AFTER UPDATE ON blackboard_entries BEGIN
		INSERT INTO blackboard_entries_fts(blackboard_entries_fts, rowid, content, category, author)
		VALUES ('delete', old.rowid, old.content, old.category, old.author);
		INSERT INTO blackboard_entries_fts(rowid, content, category, author)
		VALUES (new.rowid, new.content, new.category, new.author);
	END;
	`

	_, err := b.db.Exec(schema)
	return err
}
