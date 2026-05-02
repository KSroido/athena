package core

import (
	"github.com/google/uuid"
)

// generateUUID generates a new UUID v4
func generateUUID() string {
	return uuid.New().String()[:8]
}
