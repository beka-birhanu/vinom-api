package i

import (
	"time"
)

// Tokenizer defines methods for generating and decoding tokens.
type Tokenizer interface {
	// Generate creates a token with the given claims and expiration duration.
	Generate(claims map[string]interface{}, expTime time.Duration) (string, error)

	// Decode validates and parses a token, returning its claims.
	Decode(token string) (map[string]interface{}, error)
}
