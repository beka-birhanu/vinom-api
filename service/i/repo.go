package i

import (
	dmn "github.com/beka-birhanu/vinom-api/domain"
	"github.com/google/uuid"
)

// UserRepo defines the interface for user persistence operations.
type UserRepo interface {
	// Save inserts or updates a user in the repository.
	// If the user already exists, it updates the record. Otherwise, it creates a new one.
	Save(user *dmn.User) error

	// ByID retrieves a user by their unique ID.
	// Returns an error if the user is not found or in case of an unexpected error.
	ByID(id uuid.UUID) (*dmn.User, error)

	// ByUsername retrieves a user by their username.
	// Returns an error if the user is not found or in case of an unexpected error.
	ByUsername(username string) (*dmn.User, error)
}
