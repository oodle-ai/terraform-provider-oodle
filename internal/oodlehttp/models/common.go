package models

import "github.com/google/uuid"

// ID is a model for a unique identifier used in all alert related models.
type ID struct {
	UUID uuid.UUID
}
