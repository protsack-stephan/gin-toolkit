package httpmw

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const authTestUsername = "alex_name"
const authTestCogntoClientId = "client-id-123123123"
const authTestCogntoWrongClientId = "client-id-123123124"
const authTestIPRanges = "192.168.20.1-192.168.20.10"
const authTestIPRangesLarge = "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"

type cognitoIdentityProviderClientMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (c *cognitoIdentityProviderClientMock) GetUser(input *cognitoidentityprovider.GetUserInput) (*cognitoidentityprovider.GetUserOutput, error) {
	args := c.Called(input)

	return args.Get(0).(*cognitoidentityprovider.GetUserOutput), args.Error(1)
}

func getJWTtoken() (string, error) {
	claims := jwt.MapClaims{
		"client_id": authTestCogntoClientId,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte{})
}

func TestCognitoIPSucceed(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(IpCognitoAuth(authTestIPRangesLarge, &cognitoIdentityProviderClientMock{}, authTestCogntoClientId))
	router.GET("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(err)
	req.Header.Set("X-Forwarded-For", "192.168.20.2")
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
}

func TestCognitoIP401(t *testing.T) {
	called := false
	assert := assert.New(t)
	router := gin.New()
	router.Use(IpCognitoAuth(authTestIPRangesLarge, &cognitoIdentityProviderClientMock{}, authTestCogntoClientId))
	router.GET("/login", func(c *gin.Context) {
		called = true

		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(err)
	req.Header.Set("X-Forwarded-For", "192.168.20.20")
	router.ServeHTTP(w, req)

	assert.False(called)
	assert.Equal(http.StatusUnauthorized, w.Code)
}

func TestCognitoIpAuth(t *testing.T) {
	assert := assert.New(t)
	username := authTestUsername
	router := gin.New()
	token, err := getJWTtoken()

	assert.NoError(err)
	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &username,
			},
			nil,
		)

	router.Use(IpCognitoAuth(authTestIPRanges, srv, authTestCogntoClientId))
	router.GET("/login", func(c *gin.Context) {
		uname, _ := c.Get("username")

		assert.Equal(authTestUsername, *uname.(*string))
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
}

func TestCognitoIpAuthTokenFails(t *testing.T) {
	assert := assert.New(t)
	username := authTestUsername
	router := gin.New()
	token, err := getJWTtoken()

	assert.NoError(err)
	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &username,
			},
			nil,
		)

	router.Use(IpCognitoAuth(authTestIPRanges, srv, authTestCogntoWrongClientId))
	router.GET("/login", func(c *gin.Context) {
		uname, _ := c.Get("username")

		assert.Equal(authTestUsername, *uname.(*string))
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusUnauthorized, w.Code)
}

func TestCognitoIpAuthFails(t *testing.T) {
	assert := assert.New(t)
	called := false
	router := gin.New()
	token, err := getJWTtoken()

	assert.NoError(err)
	srv := new(cognitoIdentityProviderClientMock)
	srv.
		On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &token}).
		Return(
			&cognitoidentityprovider.GetUserOutput{},
			errors.New("token is not valid"),
		)

	router.Use(IpCognitoAuth(authTestIPRanges, srv, authTestCogntoClientId))
	router.GET("/login", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(w, req)
	assert.False(called)
	assert.Equal(http.StatusUnauthorized, w.Code)
}
