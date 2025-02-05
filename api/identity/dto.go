package identity

import "github.com/google/uuid"

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Rating   int       `json:"rating"`
	Token    string    `json:"auth_token"`
}
