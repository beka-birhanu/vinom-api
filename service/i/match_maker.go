package i

import (
	"context"

	"github.com/google/uuid"
)

type Matchmaker interface {
	Match(ctx context.Context, id uuid.UUID, rating int, latency uint) error
}
