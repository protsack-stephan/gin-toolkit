package httpmw

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

const authTestToken = "token-sg332f"
const authTestUsername = "alex_name"

type cognitoIdentityProviderClientMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (c *cognitoIdentityProviderClientMock) GetUser(input *cognitoidentityprovider.GetUserInput) (*cognitoidentityprovider.GetUserOutput, error) {
	args := c.Called(input)

	return args.Get(0).(*cognitoidentityprovider.GetUserOutput), args.Error(1)
}

func TestCognitoIPSucceed(t *testing.T) {
	assert := assert.New(t)
	ipRanges := "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"
	router := gin.New()
	router.Use(IpCognitoAuth(ipRanges, &cognitoIdentityProviderClientMock{}))
	router.GET("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	req.Header.Set("X-Forwarded-For", "192.168.20.2")
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
}

func TestCognitoIP401(t *testing.T) {
	called := false
	assert := assert.New(t)
	ipRanges := "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"
	router := gin.New()
	router.Use(IpCognitoAuth(ipRanges, &cognitoIdentityProviderClientMock{}))
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
}

func TestCognitoIpAuth(t *testing.T) {
	assert := assert.New(t)
	ipRanges := "192.168.20.1-192.168.20.10"
	token := authTestToken
	username := authTestUsername
	router := gin.New()

	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &username,
			},
			nil,
		)

	router.Use(IpCognitoAuth(ipRanges, srv))
	router.GET("/login", func(c *gin.Context) {
		uname, _ := c.Get("username")

		assert.Equal(authTestUsername, *uname.(*string))
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", "Bearer", token))
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
}

func TestCognitoIpAuthFails(t *testing.T) {
	assert := assert.New(t)
	called := false
	ipRanges := "192.168.20.1-192.168.20.10"
	token := authTestToken
	router := gin.New()

	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{},
			errors.New("token is not valid"),
		)

	router.Use(IpCognitoAuth(ipRanges, srv))
	router.GET("/login", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", "Bearer", token))
	router.ServeHTTP(w, req)
	assert.False(called)
	assert.Equal(http.StatusUnauthorized, w.Code)
}
