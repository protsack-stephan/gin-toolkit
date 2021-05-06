package httpmw

import (
	"bytes"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
)

// Range define range with start and end IPs
type Range struct {
	Start net.IP
	End   net.IP
}

// Ranges array of Range
type Ranges []Range

type authPair struct {
	value string
	user  string
}

type authPairs []authPair

func (a authPairs) searchCredential(authValue string) (string, bool) {
	if authValue == "" {
		return "", false
	}
	for _, pair := range a {
		if pair.value == authValue {
			return pair.user, true
		}
	}
	return "", false
}

// IPBasicAuth middleware for:
// * IP ranges verification in format "192.168.10.1-192.168.10.10,192.168.90.1-192.168.90.10"
// * basic authentication in format "user:pass,user2:pass2,user3:pass3"
func IPBasicAuth(ipRange string, authStorage string) gin.HandlerFunc {
	ipRanges := Ranges{}

	for _, ipRange := range strings.Split(ipRange, ",") {
		if len(ipRange) > 0 {
			ips := strings.Split(ipRange, "-")
			ipRanges = append(
				ipRanges,
				Range{
					net.ParseIP(ips[0]),
					net.ParseIP(ips[1]),
				},
			)
		}
	}

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

func checkIP(ipRange Range, ip string) bool {
	input := net.ParseIP(ip)

	return bytes.Compare(input, ipRange.Start) >= 0 && bytes.Compare(input, ipRange.End) <= 0
}

func processAccounts(accounts gin.Accounts) authPairs {
	pairs := make(authPairs, 0, len(accounts))

	if len(accounts) <= 0 {
		return pairs
	}

	for user, password := range accounts {
		if user == "" {
			continue
		}

		value := authorizationHeader(user, password)
		pairs = append(pairs, authPair{
			value: value,
			user:  user,
		})
	}
	return pairs
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password

	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}
