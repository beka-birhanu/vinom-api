package identity

import (
	"net/http"
	"strings"

	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/gin-gonic/gin"
)

const (
	// ContextUserClaims is the key used to store user claims in the Gin context.
	ContextUserClaims = "userClaims"
)

func Authoriz(ts i.Tokenizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the access token from the Authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Status(http.StatusUnauthorized) // No token found in the header.
			c.Abort()
			return
		}

		// Split the "Bearer" prefix from the token.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Status(http.StatusUnauthorized) // Malformed Authorization header.
			c.Abort()
			return
		}

		// Extract the token part.
		token := parts[1]

		// Validate the token using the barrier service.
		claims, err := ts.Decode(token)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// Attach user claims to the request context for further use.
		c.Set(ContextUserClaims, claims)
		c.Next()
	}
}
