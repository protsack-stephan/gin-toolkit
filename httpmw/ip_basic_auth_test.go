package httpmw

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicIPSucceed(t *testing.T) {
	assert := assert.New(t)
	ipRanges := "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"
	router := gin.New()
	router.Use(IPBasicAuth(ipRanges, "user:password"))
	router.GET("/login", func(c *gin.Context) {
		_, ok := c.Get(gin.AuthUserKey)

		assert.False(ok)
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	req.Header.Set("X-Forwarded-For", "192.168.20.2")
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
}

func TestBasicIP401(t *testing.T) {
	called := false
	assert := assert.New(t)
	ipRanges := "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"
	router := gin.New()
	router.Use(IPBasicAuth(ipRanges, "user:password"))
	router.GET("/login", func(c *gin.Context) {
		called = true

		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	req.Header.Set("X-Forwarded-For", "192.168.20.20")
	router.ServeHTTP(w, req)

	assert.False(called)
	assert.Equal(http.StatusUnauthorized, w.Code)
	assert.Equal("Basic realm=\"Authorization Required\"", w.Header().Get("WWW-Authenticate"))
}

func TestBasicIPAuth(t *testing.T) {
	assert := assert.New(t)
	pairs := processAccounts(gin.Accounts{
		"admin": "password",
		"foo":   "bar",
		"bar":   "foo",
	})

	assert.Len(pairs, 3)
	assert.Contains(pairs, authPair{
		user:  "bar",
		value: "Basic YmFyOmZvbw==",
	})
	assert.Contains(pairs, authPair{
		user:  "foo",
		value: "Basic Zm9vOmJhcg==",
	})
	assert.Contains(pairs, authPair{
		user:  "admin",
		value: "Basic YWRtaW46cGFzc3dvcmQ=",
	})
}

func TestBasicIPAuthFails(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(0, len(processAccounts(nil)))

	pairs := processAccounts(gin.Accounts{
		"":    "password",
		"foo": "bar",
	})

	assert.Equal(1, len(pairs))
}

func TestBasicIPAuthSearchCredential(t *testing.T) {
	assert := assert.New(t)
	pairs := processAccounts(gin.Accounts{
		"admin": "password",
		"foo":   "bar",
		"bar":   "foo",
	})

	user, found := pairs.searchCredential(authorizationHeader("admin", "password"))
	assert.Equal("admin", user)
	assert.True(found)

	user, found = pairs.searchCredential(authorizationHeader("foo", "bar"))
	assert.Equal("foo", user)
	assert.True(found)

	user, found = pairs.searchCredential(authorizationHeader("bar", "foo"))
	assert.Equal("bar", user)
	assert.True(found)

	user, found = pairs.searchCredential(authorizationHeader("admins", "password"))
	assert.Empty(user)
	assert.False(found)

	user, found = pairs.searchCredential(authorizationHeader("foo", "bar "))
	assert.Empty(user)
	assert.False(found)

	user, found = pairs.searchCredential("")
	assert.Empty(user)
	assert.False(found)
}

func TestBasicIPAuthAuthorizationHeader(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("Basic YWRtaW46cGFzc3dvcmQ=", authorizationHeader("admin", "password"))
}

func TestBasicIPAuthSucceed(t *testing.T) {
	assert := assert.New(t)
	ipRanges := "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"
	creds := "admin:password,user1:pass1"
	router := gin.New()
	router.Use(IPBasicAuth(ipRanges, creds))
	router.GET("/login", func(c *gin.Context) {
		c.String(http.StatusOK, c.MustGet(gin.AuthUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	req.Header.Set("X-Forwarded-For", "192.168.20.20")
	req.Header.Set("Authorization", authorizationHeader("user1", "pass1"))
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("user1", w.Body.String())
}

func TestBasicIPAuth401(t *testing.T) {
	called := false
	creds := "user1:pass1"
	router := gin.New()
	router.Use(IPBasicAuth("", creds))
	router.GET("/login", func(c *gin.Context) {
		called = true
		c.String(http.StatusOK, c.MustGet(gin.AuthUserKey).(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:password")))
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Basic realm=\"Authorization Required\"", w.Header().Get("WWW-Authenticate"))
}
