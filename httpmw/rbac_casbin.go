package httpmw

import (
	"errors"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// ErrNoUser indicates there is no user (or the user is nil) in the request's context.
var ErrNoUser = errors.New("missing user in request context")

// CasbinRBACAuthorizer uses a provided Casbin enforcer to implement RBAC middleware.
// This function will look up for a `CognitoUser` instance stored in the request's
// `gin.Context` using the `user` key, and will attempt to authorize the request
// using each one of the user's roles.
// If no match is made, the request will be rejected.
func CasbinRBACAuthorizer(e *casbin.Enforcer) RBACAuthorizeFunc {
	return func(c *gin.Context) (bool, error) {
		var user *CognitoUser
		if val, ok := c.Get("user"); ok && val != nil {
			user, _ = val.(*CognitoUser)
		}

		if user == nil {
			return false, ErrNoUser
		}

		for _, role := range user.GetGroups() {
			res, err := e.Enforce(role, c.Request.URL.Path, c.Request.Method)
			if err != nil {
				return false, err
			}

			if res {
				return true, nil
			}
		}

		return false, nil
	}
}
