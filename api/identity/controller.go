package identity

import (
	"net/http"

	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/gin-gonic/gin"
)

// IdentityServer handles HTTP requests related to authentication.
type IdentityServer struct {
	authService i.Authenticator
}

// NewIdentityServer creates a new AuthServer.
func NewIdentityServer(a i.Authenticator) *IdentityServer {
	return &IdentityServer{
		authService: a,
	}
}

// RegisterPublic registers public routes.
func (c *IdentityServer) RegisterPublic(route *gin.RouterGroup) {
	auth := route.Group("/auth")
	{
		auth.POST("/register", c.registerUser)
		auth.POST("/login", c.login)
	}
}

// RegisterProtected registers privileged routes.
func (c *IdentityServer) RegisterProtected(route *gin.RouterGroup) {
}

// registerUser handles user registration.
func (c *IdentityServer) registerUser(ctx *gin.Context) {
	var request AuthRequest

	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.authService.Register(request.Username, request.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{"message": "User registered successfully"}
	ctx.JSON(http.StatusCreated, response)
}

// login handles user login.
func (c *IdentityServer) login(ctx *gin.Context) {
	var request AuthRequest

	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := c.authService.SignIn(request.Username, request.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := &AuthResponse{
		ID:       user.ID.String(),
		Username: user.Username,
		Rating:   user.Rating,
		Token:    token,
	}
	ctx.JSON(http.StatusOK, response)
}
