package httpmod

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// ErrEmptyModules provided empty module list
var ErrEmptyModules = errors.New("empty modules")

// Init create modules
func Init(router *gin.Engine, modules []func() Module) error {
	if len(modules) <= 0 {
		return ErrEmptyModules
	}

	for _, getter := range modules {
		module := getter()
		group := router.Group(module.Path)

		for _, middleware := range module.Middleware {
			group.Use(middleware)
		}

		for _, route := range module.Routes {
			group.Handle(route.Method, route.Path, append(route.Middleware, route.Handler)...)
		}
	}

	return nil
}
