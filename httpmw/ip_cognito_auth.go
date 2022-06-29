package httpmw

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// JWK JSON web keys list.
type JWK struct {
	Keys []*Key `json:"keys"`
	mut  sync.Mutex
}

// Fetch get keys from the source.
func (j *JWK) Fetch(iss interface{}) error {
	if len(j.Keys) > 0 {
		return nil
	}

	j.mut.Lock()
	defer j.mut.Unlock()

	res, err := http.Get(fmt.Sprintf("%s/.well-known/jwks.json", iss))

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}

	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(j)
}

// Find key by identifier.
func (j *JWK) Find(kid string) (*Key, error) {
	for _, key := range j.Keys {
		if key.KID == kid {
			return key, nil
		}
	}

	return nil, errors.New("key not found")
}

// Key public key meta data.
type Key struct {
	Alg    string `json:"alg"`
	E      string `json:"e"`
	KID    string `json:"kid"`
	KTY    string `json:"kty"`
	N      string `json:"n"`
	Use    string `json:"use"`
	mut    sync.Mutex
	rsa256 *rsa.PublicKey
}

// RSA256 convert payload to a valid public key.
func (k *Key) RSA256() (*rsa.PublicKey, error) {
	if k.rsa256 != nil {
		return k.rsa256, nil
	}

	k.mut.Lock()
	defer k.mut.Unlock()

	edc, err := base64.RawURLEncoding.DecodeString(k.E)

	if err != nil {
		return nil, err
	}

	if len(edc) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(edc):], edc)
		edc = ndata
	}

	pub := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(edc[:])),
	}

	dcn, err := base64.RawURLEncoding.DecodeString(k.N)

	if err != nil {
		return nil, err
	}

	pub.N.SetBytes(dcn)
	k.rsa256 = pub

	return pub, nil
}

type IpCognitoParams struct {
	Srv      cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Cache    redis.Cmdable
	ClientID string
	IpRange  string
	Expire   time.Duration
	Username string
	Groups   []string
}

// CognitoClaims claims object for cognito JWT token.
type CognitoClaims struct {
	jwt.StandardClaims
	ClientID string   `json:"client_id"`
	ISS      string   `json:"iss"`
	Groups   []string `json:"cognito:groups"`
}

// IpCognitoAuth middleware for:
// * IP ranges verification in format "192.168.10.1-192.168.10.10,192.168.90.1-192.168.90.10"
// * cognito authentication through Authorization Bearer Token
// Note:
// If the expiration duration is less than one, the items in the cache never expire (by default), and must be deleted manually.
// If the cleanup interval is less than one, expired items are not deleted from the cache.
func IpCognitoAuth(p *IpCognitoParams) gin.HandlerFunc {
	ipRanges := getIpRanges(p.IpRange)
	jwk := new(JWK)

	return func(c *gin.Context) {
		if len(ipRanges) > 0 {
			for _, ipRange := range ipRanges {
				if checkIP(ipRange, c.ClientIP()) {
					c.Set("user", &CognitoUser{
						Username: p.Username,
						Groups:   p.Groups,
					})
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

		user := new(CognitoUser)

		_, err := jwt.ParseWithClaims(token, new(CognitoClaims), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)

			if !ok {
				return nil, errors.New("kid header not found")
			}

			claims, ok := token.Claims.(*CognitoClaims)

			if !ok {
				return nil, errors.New("couldn't resolve claims")
			}

			if claims.ClientID != p.ClientID {
				return nil, errors.New("incorrect client id")
			}

			if err := jwk.Fetch(claims.ISS); err != nil {
				return nil, err
			}

			key, err := jwk.Find(kid)

			if err != nil {
				return nil, err
			}

			user.SetGroups(claims.Groups)

			return key.RSA256()
		})

		if err != nil {
			log.Println(err)
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		key := fmt.Sprintf("access_token:%s", token)
		data, err := p.Cache.Get(c, key).Bytes()

		if err != nil && err != redis.Nil {
			httperr.InternalServerError(c, err.Error())
			c.Abort()
			return
		}

		if err == nil {
			if err := json.Unmarshal(data, user); err != nil {
				httperr.InternalServerError(c, err.Error())
				c.Abort()
				return
			}
		}

		if err == redis.Nil {
			res, err := p.Srv.GetUser(&cognitoidentityprovider.GetUserInput{AccessToken: &token})

			if err != nil {
				httperr.Unauthorized(c, err.Error())
				c.Abort()
				return
			}

			user.SetUsername(*res.Username)
			data, err := json.Marshal(user)

			if err != nil {
				httperr.InternalServerError(c, err.Error())
				c.Abort()
				return
			}

			p.Cache.Set(c, key, data, p.Expire)
		}

		c.Set("user", user)
		c.Next()
	}
}
