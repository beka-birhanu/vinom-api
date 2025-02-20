package matchmaking

import (
	"context"
	"fmt"

	"github.com/beka-birhanu/vinom-api/service/i"
	general_i "github.com/beka-birhanu/vinom-common/interfaces/general"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
)

type clientAdapter struct {
	client MatchmakingClient
	logger general_i.Logger
}

func NewClient(cc grpc.ClientConnInterface, logger general_i.Logger) (i.Matchmaker, error) {
	client := NewMatchmakingClient(cc)
	return &clientAdapter{
		client: client,
		logger: logger,
	}, nil
}

// Match implements i.MatchMakingRequester.
func (c *clientAdapter) Match(ctx context.Context, id uuid.UUID, rating int, latency uint) error {
	request := &MatchRequest{
		ID:      id.String(),
		Rating:  int32(rating),
		Latency: int32(latency),
	}

	c.logger.Info(fmt.Sprintf("sending match request for player: %s", id))
	_, err := c.client.Match(ctx, request)
	if err != nil {
		c.logger.Error(fmt.Sprintf("match request failed for player %s: %s", id, err))
		return err
	}

	c.logger.Info(fmt.Sprintf("match request success for player %s", id))
	return nil
}
