package httpmiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// NotFound handler for not found routes
func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, httperr.NotFound)
	}
}
