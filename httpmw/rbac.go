package httpmw

import (
	"github.com/gin-gonic/gin"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// RBACAuthorizeFunc is the type alias for a RBAC Authorize function signature.
type RBACAuthorizeFunc func(*gin.Context) (bool, error)

// RBAC implements RBAC using the provided authorizer function.
func RBAC(fn RBACAuthorizeFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, err := fn(c)

		if err != nil {
			httperr.InternalServerError(c, err.Error())
			c.Abort()
			return
		}

		if !ok {
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
