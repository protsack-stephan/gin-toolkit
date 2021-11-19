package httpmw

import (
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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

		realm := "Authorization Required"
		authHeader := strings.Split(c.GetHeader("Authorization"), "Bearer ")

		if len(authHeader) != 2 || authHeader[1] == "" {
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token := authHeader[1]
		res, err := svc.GetUser(&cognitoidentityprovider.GetUserInput{AccessToken: &token})

		if err != nil {
			log.Println(err)
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("username", res.Username)
		c.Next()
	}
}
