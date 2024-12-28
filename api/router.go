package api

import (
	"github.com/beka-birhanu/vinom-api/api/i"
	"github.com/gin-gonic/gin"
)

// Router manages the HTTP server and its dependencies,
// including controllers and JWT authentication.
type Router struct {
	addr                    string
	baseURL                 string
	controllers             []i.Controller
	authorizationMiddleware gin.HandlerFunc
}

// Config holds configuration settings for creating a new Router instance.
type Config struct {
	Addr                    string // Address to listen on
	BaseURL                 string // Base URL for API routes
	Controllers             []i.Controller
	AuthorizationMiddleware gin.HandlerFunc
}

// NewRouter creates a new Router instance with the given configuration.
// It initializes the router with address, base URL, controllers, and JWT service.
func NewRouter(config Config) *Router {
	return &Router{
		addr:                    config.Addr,
		baseURL:                 config.BaseURL,
		controllers:             config.Controllers,
		authorizationMiddleware: config.AuthorizationMiddleware,
	}
}

// Run starts the HTTP server and sets up routes with different access levels.
//
// Routes are grouped and managed under the base URL, with the following access levels:
// - Public routes: No authentication required.
// - Protected routes: Authentication required.
func (r *Router) Run() error {
	gin.ForceConsoleColor()
	router := gin.Default()

	// Setting up routes under baseURL
	api := router.Group(r.baseURL)

	{
		// Public routes (accessible without authentication)
		publicRoutes := api.Group("/v1")
		{
			for _, c := range r.controllers {
				c.RegisterPublic(publicRoutes)
			}
		}

		// Protected routes (authentication required)
		protectedRoutes := api.Group("/v1")
		protectedRoutes.Use(r.authorizationMiddleware)
		{
			for _, c := range r.controllers {
				c.RegisterProtected(protectedRoutes)
			}
		}
	}

	return router.Run(r.addr)
}
