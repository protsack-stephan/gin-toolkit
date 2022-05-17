package httpmw

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogFormatter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	t.Run("test log formatting function", func(_ *testing.T) {
		const apiPath = "/test"
		const testIP = "0.0.0.0"

		t.Run("test without cognito user middleware", func(_ *testing.T) {
			e := gin.New()
			out := new(bytes.Buffer)

			e.Use(gin.LoggerWithConfig(gin.LoggerConfig{
				Formatter: LogFormatter,
				Output:    out,
			}))

			e.GET(apiPath, func(c *gin.Context) {
				c.Status(http.StatusTeapot)
			})

			req, err := http.NewRequest(http.MethodGet, apiPath, nil)
			assert.NoError(err)
			req.Header.Add("X-Forwarded-For", testIP)

			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)

			entry := new(logEntry)
			assert.NoError(json.Unmarshal(out.Bytes(), entry))

			assert.NotEmpty(entry.ResponseTime)
			assert.NotEmpty(entry.Latency)
			assert.NotZero(entry.BodySize)
			assert.Equal(http.MethodGet, entry.Method)
			assert.Equal(testIP, entry.IP)
			assert.Equal(apiPath, entry.Path)
			assert.Equal(w.Code, entry.Status)
		})

		t.Run("test with cognito user middleware", func(_ *testing.T) {
			e := gin.New()
			out := new(bytes.Buffer)

			testUser := &CognitoUser{
				Username: "test_username",
				Groups:   []string{"group_1"},
			}

			e.Use(func(c *gin.Context) {
				c.Set("user", testUser)
			})

			e.Use(gin.LoggerWithConfig(gin.LoggerConfig{
				Formatter: LogFormatter,
				Output:    out,
			}))

			e.GET(apiPath, func(c *gin.Context) {
				c.Status(http.StatusTeapot)
			})

			req, err := http.NewRequest(http.MethodGet, apiPath, nil)
			assert.NoError(err)
			req.Header.Add("X-Forwarded-For", testIP)

			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)

			entry := new(logEntry)
			assert.NoError(json.Unmarshal(out.Bytes(), entry))

			assert.Equal(testUser.Username, entry.Username)
			assert.ElementsMatch(testUser.Groups, entry.UserGroups)
		})
	})
}
