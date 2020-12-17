package httpmod

import "github.com/gin-gonic/gin"

// Module struct to represent module
type Module struct {
	Path       string
	Middleware []gin.HandlerFunc
	Routes     []Route
}
