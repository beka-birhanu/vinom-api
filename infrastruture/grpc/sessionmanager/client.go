package grpc_sessionmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/beka-birhanu/vinom-api/service/i"
	general_i "github.com/beka-birhanu/vinom-common/interfaces/general"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
)

type clientAdapter struct {
	client     SessionClient
	logger     general_i.Logger
	rpcTimeout time.Duration
}

func NewClient(cc grpc.ClientConnInterface, logger general_i.Logger, rt time.Duration) (i.GameSessionManager, error) {
	client := NewSessionClient(cc)
	return &clientAdapter{
		client:     client,
		logger:     logger,
		rpcTimeout: rt,
	}, nil
}

// SessionInfo implements i.GameSessionInfoRequester.
func (c *clientAdapter) SessionInfo(ctx context.Context, id uuid.UUID) ([]byte, string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.rpcTimeout)
	defer cancel()

	request := &SessionInfoRequest{
		PlayerID: id.String(),
	}

	c.logger.Info(fmt.Sprintf("sending session info request for player: %s", id))
	res, err := c.client.SessionInfo(timeoutCtx, request)
	if err != nil {
		c.logger.Error(fmt.Sprintf("session info request failed for player %s: %s", id, err))
		return nil, "", err
	}

	c.logger.Info(fmt.Sprintf("session info request success for player %s", id))
	return []byte(res.GetServerPubKey()), res.GetServerAddr(), nil
}
