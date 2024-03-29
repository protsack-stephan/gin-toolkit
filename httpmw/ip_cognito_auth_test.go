package httpmw

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const authTestKID = "pqZ9xSMr5rtwrPG2LRM9v"
const authTestUsername = "john_doe"
const authTestUserGroup = "admin"
const authTestClientID = "jN4Ag4CEL2TQtrqk"

var authTestUserGroups = []string{authTestUserGroup}

const authTestWrongClientID = "VnFAL5ke9hK8v6bT"
const authTestIPRanges = "192.168.20.1-192.168.20.10"
const authTestIPRangesLarge = "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"

var authTestPrivateKey *rsa.PrivateKey

func init() {
	var err error

	if authTestPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		log.Panic(err)
	}
}

type cognitoIdentityProviderClientMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (c *cognitoIdentityProviderClientMock) GetUser(input *cognitoidentityprovider.GetUserInput) (*cognitoidentityprovider.GetUserOutput, error) {
	args := c.Called(input)
	return args.Get(0).(*cognitoidentityprovider.GetUserOutput), args.Error(1)
}

func createJWKServer() *httptest.Server {
	router := gin.New()

	router.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, JWK{
			Keys: []*Key{
				{
					KID: authTestKID,
					E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(authTestPrivateKey.PublicKey.E)).Bytes()),
					N:   base64.RawURLEncoding.EncodeToString(authTestPrivateKey.PublicKey.N.Bytes()),
				},
			},
		})
	})

	return httptest.NewServer(router)
}

func getJWTToken(iss string) (string, error) {
	token := jwt.
		NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"client_id":      authTestClientID,
			"iss":            iss,
			"cognito:groups": []string{authTestUserGroup},
		})

	token.Header["kid"] = authTestKID

	return token.
		SignedString(authTestPrivateKey)
}

func TestCognitoIPSucceed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	router := gin.New()
	cmdable := redis.NewClient(&redis.Options{})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      &cognitoIdentityProviderClientMock{},
		Cache:    cmdable,
		ClientID: authTestClientID,
		IpRange:  authTestIPRangesLarge,
		Expire:   time.Minute * 1,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		user, _ := c.Get("user")
		assert.Equal(authTestUsername, user.(*CognitoUser).GetUsername())
		assert.Contains(user.(*CognitoUser).GetGroups(), authTestUserGroup)
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("X-Forwarded-For", "192.168.20.2")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
}

func TestCognitoIP401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)
	called := false

	router := gin.New()
	cmdable := redis.NewClient(&redis.Options{})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      &cognitoIdentityProviderClientMock{},
		Cache:    cmdable,
		ClientID: authTestClientID,
		IpRange:  authTestIPRangesLarge,
		Expire:   time.Minute * 1,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		called = true
		_, exists := c.Get("user")
		assert.Equal(exists, false)
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("X-Forwarded-For", "192.168.20.20")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.False(called)
	assert.Equal(http.StatusUnauthorized, w.Code)
}

func TestCognitoIpAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	jwk := createJWKServer()
	defer jwk.Close()

	token, err := getJWTToken(jwk.URL)
	assert.NoError(err)

	username := authTestUsername
	srv := new(cognitoIdentityProviderClientMock)

	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &username,
			},
			nil,
		)

	router := gin.New()
	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      srv,
		Cache:    cmdable,
		ClientID: authTestClientID,
		IpRange:  authTestIPRanges,
		Expire:   time.Minute * 1,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		user, _ := c.Get("user")
		assert.Equal(authTestUsername, user.(*CognitoUser).GetUsername())
		assert.Contains(user.(*CognitoUser).GetGroups(), authTestUserGroup)
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
	router.ServeHTTP(w, req)
}

func TestCognitoIpAuthTokenFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	jwk := createJWKServer()
	defer jwk.Close()

	token, err := getJWTToken(jwk.URL)
	assert.NoError(err)

	username := authTestUsername
	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &username,
			},
			nil,
		)

	router := gin.New()
	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      srv,
		Cache:    cmdable,
		ClientID: authTestWrongClientID,
		IpRange:  authTestIPRanges,
		Expire:   time.Minute * 1,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		user, _ := c.Get("user")
		assert.Equal(authTestUsername, user.(*CognitoUser).GetUsername())
		assert.Contains(user.(*CognitoUser).GetGroups(), authTestUserGroup)
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusUnauthorized, w.Code)
}

func TestCognitoIpAuthFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)
	called := false

	jwk := createJWKServer()
	defer jwk.Close()

	token, err := getJWTToken(jwk.URL)
	assert.NoError(err)

	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{},
			errors.New("token is not valid"),
		)

	router := gin.New()
	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      srv,
		Cache:    cmdable,
		ClientID: authTestClientID,
		IpRange:  authTestIPRanges,
		Expire:   time.Minute * 1,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.False(called)
	assert.Equal(http.StatusUnauthorized, w.Code)
}

func TestCognitoIpAuthCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	jwk := createJWKServer()
	defer jwk.Close()

	token, err := getJWTToken(jwk.URL)
	assert.NoError(err)

	username := authTestUsername
	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &username,
			},
			nil,
		)

	router := gin.New()
	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      srv,
		Cache:    cmdable,
		ClientID: authTestClientID,
		IpRange:  authTestIPRanges,
		Expire:   time.Minute * 19,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		user, _ := c.Get("user")
		fmt.Println(user.(*CognitoUser))
		assert.Equal(authTestUsername, user.(*CognitoUser).GetUsername())
		assert.Contains(user.(*CognitoUser).GetGroups(), authTestUserGroup)
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()

	for i := 0; i < 10; i++ {
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)
	}

	srv.AssertNumberOfCalls(t, "GetUser", 1)
}

func TestCognitoIpAuthCacheExpire(t *testing.T) {
	expire := time.Millisecond
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	jwk := createJWKServer()
	defer jwk.Close()

	token, err := getJWTToken(jwk.URL)
	assert.NoError(err)

	username := authTestUsername
	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &username,
			},
			nil,
		)

	router := gin.New()
	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      srv,
		Cache:    cmdable,
		ClientID: authTestClientID,
		IpRange:  authTestIPRanges,
		Expire:   expire,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		user, _ := c.Get("user")
		assert.Equal(authTestUsername, user.(*CognitoUser).GetUsername())
		assert.Contains(user.(*CognitoUser).GetGroups(), authTestUserGroup)
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()

	for i := 0; i < 2; i++ {
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)
		mr.FastForward(expire)
	}

	srv.AssertNumberOfCalls(t, "GetUser", 2)
}

func TestCognitoIpAuthCacheUnreachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	jwk := createJWKServer()
	defer jwk.Close()

	token, err := getJWTToken(jwk.URL)
	assert.NoError(err)

	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(nil, nil)

	router := gin.New()
	cmdable := redis.NewClient(&redis.Options{})
	router.Use(IpCognitoAuth(&IpCognitoParams{
		Srv:      srv,
		Cache:    cmdable,
		ClientID: authTestClientID,
		IpRange:  authTestIPRanges,
		Expire:   time.Second * 1,
		User: &CognitoUser{
			Username: authTestUsername,
			Groups:   authTestUserGroups,
		},
	}))
	router.GET("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusInternalServerError, w.Code)

	srv.AssertNumberOfCalls(t, "GetUser", 0)
}
