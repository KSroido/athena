package blackboard

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ksroido/athena/internal/db"
)

// FactManager handles fact grading and verification logic
type FactManager struct {
	board *Board
}

// NewFactManager creates a new FactManager
func NewFactManager(board *Board) *FactManager {
	return &FactManager{board: board}
}

// CertaintyDecayDays is the default number of days before a "certain" fact
// is automatically downgraded to "pending_verification"
const CertaintyDecayDays = 7

// VerifyEntry upgrades a pending_verification entry to certain (requires cross-validation)
func (fm *FactManager) VerifyEntry(entryID string, validatorID string) error {
	entry, err := fm.getEntryByID(entryID)
	if err != nil {
		return err
	}

	if entry.Certainty != CertaintyPendingVerification {
		return fmt.Errorf("entry %s is not pending verification (current: %s)", entryID, entry.Certainty)
	}

	// Upgrade to certain
	entry.Certainty = CertaintyCertain
	now := time.Now()
	entry.LastVerifiedAt = &now
	entry.UpdatedAt = time.Now()

	return fm.board.WriteEntrySync(entry)
}

// DegradeEntry downgrades a certain fact to pending_verification or conjecture
func (fm *FactManager) DegradeEntry(entryID string, newCertainty string) error {
	entry, err := fm.getEntryByID(entryID)
	if err != nil {
		return err
	}

	entry.Certainty = newCertainty
	entry.UpdatedAt = time.Now()

	return fm.board.WriteEntrySync(entry)
}

// CheckDecay scans for "certain" facts that haven't been verified recently
// and returns entries that should be downgraded
func (fm *FactManager) CheckDecay() ([]*EntryToDegrade, error) {
	cutoff := time.Now().AddDate(0, 0, -CertaintyDecayDays)

	entries, err := fm.board.ReadEntries("", 1000, 0)
	if err != nil {
		return nil, err
	}

	var toDegrade []*EntryToDegrade
	for _, e := range entries {
		if e.Certainty == CertaintyCertain && e.LastVerifiedAt != nil {
			if e.LastVerifiedAt.Before(cutoff) {
				toDegrade = append(toDegrade, &EntryToDegrade{
					EntryID:    e.ID,
					Content:    e.Content,
					Author:     e.Author,
					LastVerify: *e.LastVerifiedAt,
				})
			}
		}
	}

	return toDegrade, nil
}

// getEntryByID retrieves a single entry by ID
func (fm *FactManager) getEntryByID(entryID string) (*db.BlackboardEntry, error) {
	// Query the board database directly
	fm.board.mu.RLock()
	defer fm.board.mu.RUnlock()

	row := fm.board.db.QueryRow(`
		SELECT id, project_id, category, content, certainty, author,
		       confidence_score, reasoning, last_verified_at, created_at, updated_at
		FROM blackboard_entries WHERE id = ?
	`, entryID)

	e := &db.BlackboardEntry{}
	var lastVerified sql.NullTime
	var confidenceScore sql.NullInt64

	err := row.Scan(&e.ID, &e.ProjectID, &e.Category, &e.Content, &e.Certainty,
		&e.Author, &confidenceScore, &e.Reasoning, &lastVerified,
		&e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entry %s not found", entryID)
	} else if err != nil {
		return nil, fmt.Errorf("query entry: %w", err)
	}

	if confidenceScore.Valid {
		score := int(confidenceScore.Int64)
		e.ConfidenceScore = &score
	}
	if lastVerified.Valid {
		e.LastVerifiedAt = &lastVerified.Time
	}

	return e, nil
}

// EntryToDegrade represents an entry that should be downgraded
type EntryToDegrade struct {
	EntryID    string    `json:"entry_id"`
	Content    string    `json:"content"`
	Author     string    `json:"author"`
	LastVerify time.Time `json:"last_verified_at"`
}
