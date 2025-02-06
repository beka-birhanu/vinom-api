// Package gameapi provides structures and utilities for managing game match requests and responses.
package gameapi

import (
	"github.com/google/uuid"
)

// MatchRequest represents a request to create a new game match.
type MatchRequest struct {
	ID     uuid.UUID `json:"id" binding:"required"`
	SentAt int64     `json:"sent_at" binding:"required"`
}

// MatchInfoResponse represents the response containing information about a specific match.
type MatchInfoResponse struct {
	SocketPubKey []byte `json:"socket_pubkey"`
	SocketAddr   string `json:"socket_addr"`
}
