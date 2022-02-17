package httpmw

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// RBACAuthorizeFunc is the type alias for a RBAC Authorize function signature.
type RBACAuthorizeFunc func(*gin.Context) (bool, error)

// RBAC implements RBAC using the provided authorizer function.
func RBAC(fn RBACAuthorizeFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authorized, err := fn(c); err != nil || !authorized {
			if err != nil {
				log.Println(err)
			}
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
