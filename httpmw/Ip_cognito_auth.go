package httpmw

import (
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/protsack-stephan/gin-toolkit/httperr"
	"log"
	"strings"
)

// IpCognitoAuth middleware for:
// * IP ranges verification in format "192.168.10.1-192.168.10.10,192.168.90.1-192.168.90.10"
// * cognito authentication through Authorization Bearer Token
func IpCognitoAuth(ipRange string, svc cognitoidentityprovideriface.CognitoIdentityProviderAPI) gin.HandlerFunc {
	ipRanges := getIpRanges(ipRange)

	return func(c *gin.Context) {
		if len(ipRanges) > 0 {
			for _, ipRange := range ipRanges {
				if checkIP(ipRange, c.ClientIP()) {
					return
				}
			}
		}

		token := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", 1)

		if len(token) <= 0 {
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		res, err := svc.GetUser(&cognitoidentityprovider.GetUserInput{AccessToken: &token})

		if err != nil {
			log.Println(err)
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		c.Set("username", res.Username)
		c.Next()
	}
}
