package httpmw

import (
	"github.com/gin-gonic/gin"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// NotFound handler for not found routes
func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		httperr.NotFound(c)
	}
}
