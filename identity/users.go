package identity

import (
	"errors"
	"regexp"

	"github.com/google/uuid"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordStrengthScore = 3

	usernamePattern   = `^[a-zA-Z0-9_]+$` // Alphanumeric with underscores
	minUsernameLength = 3
	maxUsernameLength = 20

	defautlRating = 1400
)

var (
	usernameRegex = regexp.MustCompile(usernamePattern)
)

// User represents the BSON version of the User for database storage.
type User struct {
	ID           uuid.UUID `bson:"_id"`
	Username     string    `bson:"username"`
	PasswordHash string    `bson:"passwordHash"`
	Rating       int       `bson:"rating"`
}

// UserConfig holds parameters for creating a User with an existing password hash.
type UserConfig struct {
	ID            uuid.UUID
	Username      string
	PlainPassword string
}

// New creates a new User with the provided configuration.
func NewUser(config UserConfig) (*User, error) {
	if err := validateUsername(config.Username); err != nil {
		return nil, err
	}

	if err := validatePassword(config.PlainPassword); err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(config.PlainPassword)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           config.ID,
		Username:     config.Username,
		PasswordHash: passwordHash,
		Rating:       defautlRating,
	}, nil
}

// VerifyPassword verifies if the given password matches the stored hash.
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// validateUsername validates the username.
func validateUsername(username string) error {
	if len(username) < minUsernameLength {
		return errors.New("username too short")
	}
	if len(username) > maxUsernameLength {
		return errors.New("username too long")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("Invalid username format")
	}
	return nil
}

// validatePassword checks the strength of the password.
func validatePassword(password string) error {
	result := zxcvbn.PasswordStrength(password, nil)
	if result.Score < minPasswordStrengthScore {
		return errors.New("week password")
	}
	return nil
}

// hashPassword generates a bcrypt hash for the given password.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
