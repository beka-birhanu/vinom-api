package identity

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Token    string `json:"auth_token"`
}
