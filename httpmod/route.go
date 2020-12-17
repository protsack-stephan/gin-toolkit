package httpmod

import (
	"github.com/gin-gonic/gin"
)

// Route struct to represent single route
type Route struct {
	Path       string
	Method     string
	Middleware []gin.HandlerFunc
	Handler    func(c *gin.Context)
}
