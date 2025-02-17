package i

import (
	"context"

	"github.com/google/uuid"
)

// GameSessionManager manages game sessions and provides session-related information.
type GameSessionManager interface {
	// SessionInfo returns the public key, socket address.
	SessionInfo(context.Context, uuid.UUID) ([]byte, string, error)
}
