package token

import (
	"errors"
	"time"

	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/dgrijalva/jwt-go"
)

// JwtService handles JWT operations.
// Implements ijwt.JwtService.
type JwtService struct {
	secretKey string
	issuer    string
}

// New creates a new JWT Service with the provided configuration.
func NewJwtService(secretKey, issuer string) i.Tokenizer {
	return &JwtService{
		secretKey: secretKey,
		issuer:    issuer,
	}
}

// Generate creates a JWT for the given claims.
func (s *JwtService) Generate(claims map[string]interface{}, expTime time.Duration) (string, error) {
	expirationTime := time.Now().UTC().Add(expTime).Unix()
	jwtClaims := jwt.MapClaims{
		"exp": expirationTime,
	}
	for key, val := range claims {
		jwtClaims[key] = val
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString([]byte(s.secretKey))
}

// Decode parses and validates a JWT, returning the claims if valid.
func (s *JwtService) Decode(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, s.getSigningKey)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// getSigningKey returns the signing key for token validation.
func (s *JwtService) getSigningKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return []byte(s.secretKey), nil
}
