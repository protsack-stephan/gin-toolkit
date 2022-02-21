package httpmw

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`

const casbinPolicy = `
p, free-user, /free, GET
p, trial-user, /trial, GET
p, paid-user, /paid, GET
g, trial-user, free-user
g, paid-user, trial-user
`

var testPaths = []string{"/free", "/trial", "/paid"}

func setupCasbinRBACMWUser(groups ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := new(CognitoUser)
		user.SetUsername("username")
		user.SetGroups(groups)

		c.Set("user", user)
	}
}

func getCasbinRBACTestingRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()

	// Setup middlewares
	for _, mw := range middlewares {
		r.Use(mw)
	}

	// Setup handler for each path
	for _, p := range testPaths {
		r.GET(p, func(c *gin.Context) {
			c.Status(http.StatusTeapot)
		})
	}

	return r
}

func TestRBACCasbinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assert := assert.New(t)

	t.Run("test Casbin RBAC implementation", func(_ *testing.T) {
		var err error
		var model, policy *os.File

		model, err = ioutil.TempFile("/tmp", "model")
		assert.NoError(err)

		_, err = model.Write([]byte(casbinModel))
		assert.NoError(err)

		policy, err = ioutil.TempFile("/tmp", "policy")
		assert.NoError(err)

		_, err = policy.Write([]byte(casbinPolicy))
		assert.NoError(err)

		e, err := casbin.NewEnforcer(model.Name(), policy.Name())
		assert.NoError(err)
		e.EnableLog(true)

		t.Run("test free user policies", func(_ *testing.T) {
			router := getCasbinRBACTestingRouter(
				setupCasbinRBACMWUser("free-user"),
				RBAC(CasbinRBACAuthorizeFunc(e)),
			)

			// Allow user with free-user group to consume /free endpoint
			req, err := http.NewRequest(http.MethodGet, "/free", nil)
			assert.NoError(err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)

			// Deny user with free-user group to consume /trial endpoint
			req, err = http.NewRequest(http.MethodGet, "/trial", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusUnauthorized, w.Code)

			// Deny user with free-user group to consume /paid endpoint
			req, err = http.NewRequest(http.MethodGet, "/paid", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusUnauthorized, w.Code)
		})

		t.Run("test trial user policies", func(_ *testing.T) {
			router := getCasbinRBACTestingRouter(
				setupCasbinRBACMWUser("trial-user"),
				RBAC(CasbinRBACAuthorizeFunc(e)),
			)

			// Allow user with trial-user group to consume /free endpoint
			req, err := http.NewRequest(http.MethodGet, "/free", nil)
			assert.NoError(err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)

			// Allow user with trial-user group to consume /trial endpoint
			req, err = http.NewRequest(http.MethodGet, "/trial", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)

			// Deny user with trial-user group to consume /paid endpoint
			req, err = http.NewRequest(http.MethodGet, "/paid", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusUnauthorized, w.Code)
		})

		t.Run("test paid user policies", func(_ *testing.T) {
			router := getCasbinRBACTestingRouter(
				setupCasbinRBACMWUser("paid-user"),
				RBAC(CasbinRBACAuthorizeFunc(e)),
			)

			// Allow user with paid-user group to consume /free endpoint
			req, err := http.NewRequest(http.MethodGet, "/free", nil)
			assert.NoError(err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)

			// Allow user with paid-user group to consume /trial endpoint
			req, err = http.NewRequest(http.MethodGet, "/trial", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)

			// Allow user with paid-user group to consume /paid endpoint
			req, err = http.NewRequest(http.MethodGet, "/paid", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusTeapot, w.Code)
		})

		t.Run("test user without groups", func(_ *testing.T) {
			router := getCasbinRBACTestingRouter(
				setupCasbinRBACMWUser(),
				RBAC(CasbinRBACAuthorizeFunc(e)),
			)

			// Deny user without groups to consume /free endpoint
			req, err := http.NewRequest(http.MethodGet, "/free", nil)
			assert.NoError(err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusUnauthorized, w.Code)

			// Deny user without groups to consume /trial endpoint
			req, err = http.NewRequest(http.MethodGet, "/trial", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusUnauthorized, w.Code)

			// Deny user without groups to consume /paid endpoint
			req, err = http.NewRequest(http.MethodGet, "/paid", nil)
			assert.NoError(err)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(http.StatusUnauthorized, w.Code)
		})
	})
}
