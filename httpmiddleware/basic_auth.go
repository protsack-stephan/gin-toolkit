package httpmiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// BasicAuth middleware for basic authentication in format "user:pass,user2:pass2,user3:pass3"
func BasicAuth(storage string) gin.HandlerFunc {
	users := gin.Accounts{}

	for _, account := range strings.Split(storage, ",") {
		if len(account) > 0 {
			cred := strings.Split(account, ":")
			users[cred[0]] = cred[1]
		}
	}

	return gin.BasicAuth(users)
}
