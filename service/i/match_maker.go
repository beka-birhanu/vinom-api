package i

import (
	"context"

	"github.com/google/uuid"
)

type Matchmaker interface {
	PushToQueue(ctx context.Context, id uuid.UUID, rating int, latency uint) error
	SetMatchHandler(func([]uuid.UUID))
}
