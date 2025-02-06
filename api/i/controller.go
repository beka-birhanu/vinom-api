package i

import "github.com/gin-gonic/gin"

type Controller interface {
	RegisterPublic(*gin.RouterGroup)
	RegisterProtected(*gin.RouterGroup)
}
