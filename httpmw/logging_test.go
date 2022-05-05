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
		t.Run("test without cognito user middleware", func(_ *testing.T) {
			e := gin.New()
			out := new(bytes.Buffer)
			apiPath := "/test"
			testIP := "0.0.0.0"

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
			err = json.Unmarshal(out.Bytes(), entry)
			assert.NoError(err)

			assert.NotEmpty(entry.RequestTime)
			assert.Equal(testIP, entry.Ip)
			assert.Equal(apiPath, entry.Path)
			assert.Equal(w.Code, entry.Status)
		})
	})

	t.Run("test with cognito user middleware", func(_ *testing.T) {
		e := gin.New()
		out := new(bytes.Buffer)
		apiPath := "/test"
		testIP := "0.0.0.0"

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
		err = json.Unmarshal(out.Bytes(), entry)
		assert.NoError(err)

		assert.Equal(testUser.Username, entry.Username)
		assert.ElementsMatch(testUser.Groups, entry.UserGroups)
	})
}
