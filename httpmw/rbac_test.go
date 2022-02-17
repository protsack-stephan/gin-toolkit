package httpmw

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	t.Run("test RBAC middleware", func(_ *testing.T) {
		handler := func(c *gin.Context) { c.Status(http.StatusTeapot) }
		t.Run("test access granted", func(_ *testing.T) {
			router := gin.New()
			router.Use(RBAC(func(c *gin.Context) (bool, error) {
				return true, nil
			}))
			router.GET("/test", handler)

			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			assert.NoError(err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(http.StatusTeapot, w.Code)
		})

		t.Run("test access denied", func(t *testing.T) {
			router := gin.New()
			router.Use(RBAC(func(c *gin.Context) (bool, error) {
				return false, nil
			}))
			router.GET("/test", handler)

			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			assert.NoError(err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(http.StatusUnauthorized, w.Code)
		})

		t.Run("test access denied with error", func(t *testing.T) {
			router := gin.New()
			router.Use(RBAC(func(c *gin.Context) (bool, error) {
				return true, errors.New("test error")
			}))
			router.GET("/test", handler)

			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			assert.NoError(err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(http.StatusUnauthorized, w.Code)
		})
	})
}
