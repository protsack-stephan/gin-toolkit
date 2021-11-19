package httpmw

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type authPair struct {
	value string
	user  string
}

type authPairs []authPair

func (a authPairs) searchCredential(authValue string) (string, bool) {
	if authValue != "" {
		for _, pair := range a {
			if pair.value == authValue {
				return pair.user, true
			}
		}
	}

	return "", false
}

// IPBasicAuth middleware for:
// * IP ranges verification in format "192.168.10.1-192.168.10.10,192.168.90.1-192.168.90.10"
// * basic authentication in format "user:pass,user2:pass2,user3:pass3"
func IPBasicAuth(ipRange string, authStorage string) gin.HandlerFunc {
	ipRanges := getIpRanges(ipRange)
	accounts := gin.Accounts{}

	for _, account := range strings.Split(authStorage, ",") {
		if len(account) > 0 {
			cred := strings.Split(account, ":")
			accounts[cred[0]] = cred[1]
		}
	}

	pairs := processAccounts(accounts)

	return func(c *gin.Context) {
		if len(ipRanges) > 0 {
			for _, ipRange := range ipRanges {
				if checkIP(ipRange, c.ClientIP()) {
					return
				}
			}
		}

		if len(pairs) > 0 {
			user, found := pairs.searchCredential(c.Request.Header.Get("Authorization"))

			if !found {
				c.Header("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			// The user credentials was found, set user's id to key AuthUserKey in this context,
			// the user's id can be read later using c.MustGet(gin.AuthUserKey).
			c.Set(gin.AuthUserKey, user)
		}
	}
}

func processAccounts(accounts gin.Accounts) authPairs {
	pairs := make(authPairs, 0, len(accounts))

	for user, password := range accounts {
		if user != "" {
			pairs = append(pairs, authPair{
				value: authorizationHeader(user, password),
				user:  user,
			})
		}
	}

	return pairs
}

func authorizationHeader(user, password string) string {
	base := fmt.Sprintf("%s:%s", user, password)

	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(base)))
}
