package token

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJwtService(t *testing.T) {
	// Setup
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Error generating random bytes: %v", err)
	}
	secretKey := base64.URLEncoding.EncodeToString(bytes)
	issuer := "testIssuer"

	svc := NewJwtService(secretKey, issuer)

	t.Run("Generate and Decode valid token", func(t *testing.T) {
		claims := map[string]interface{}{
			"user_id": 12345,
			"role":    "admin",
		}
		expDuration := time.Minute * 5

		// Generate token
		token, err := svc.Generate(claims, expDuration)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Decode token
		_, err = svc.Decode(token)
		assert.NoError(t, err)
	})

	t.Run("Decode invalid token", func(t *testing.T) {
		invalidToken := "invalidTokenString"

		// Decode should fail
		_, err := svc.Decode(invalidToken)
		assert.Error(t, err)
	})

	t.Run("Decode expired token", func(t *testing.T) {
		claims := map[string]interface{}{
			"user_id": 12345,
			"role":    "admin",
		}
		expDuration := -time.Minute // Expired token

		// Generate expired token
		token, err := svc.Generate(claims, expDuration)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Decode should fail
		_, err = svc.Decode(token)
		assert.Error(t, err)
	})

	t.Run("Generate token with empty claims", func(t *testing.T) {
		emptyClaims := map[string]interface{}{}
		expDuration := time.Minute * 5

		// Generate token
		token, err := svc.Generate(emptyClaims, expDuration)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Decode token
		decodedClaims, err := svc.Decode(token)
		assert.NoError(t, err)
		assert.Empty(t, decodedClaims["user_id"])
		assert.Empty(t, decodedClaims["role"])
	})
}
