package identity

import (
	"net/http"

	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// IdentityServer handles HTTP requests related to authentication.
type IdentityServer struct {
	authService i.Authenticator
}

// IdentityServerConfig contains configuration options for IdentityServer.
type IdentityServerConfig struct {
	MongoClient *mongo.Client
	DBName      string
	JWTSecret   string
	JWTIssuer   string
}

// NewIdentityServer creates a new AuthServer.
func NewIdentityServer(config IdentityServerConfig) *IdentityServer {
	return &IdentityServer{}
}

// RegisterPublic registers public routes.
func (c *IdentityServer) RegisterPublic(route *gin.RouterGroup) {
	auth := route.Group("/auth")
	{
		auth.POST("/register", c.registerUser)
		auth.POST("/login", c.login)
	}
}

// RegisterPrivileged registers privileged routes.
func (c *IdentityServer) RegisterPrivileged(route *gin.RouterGroup) {
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
