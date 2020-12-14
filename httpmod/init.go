package httpmod

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// ErrEmptyModules provided empty module list
var ErrEmptyModules = errors.New("empty modules")

// Init create modules
func Init(router *gin.Engine, modules []Module) error {
	if len(modules) <= 0 {
		return ErrEmptyModules
	}

	for _, module := range modules {
		group := router.Group(module.Path)

		for _, middleware := range module.Middleware {
			group.Use(middleware())
		}

		for _, route := range module.Routes {
			handlers := []gin.HandlerFunc{}

			for _, middleware := range route.Middleware {
				handlers = append(handlers, middleware())
			}

			group.Handle(route.Method, route.Path, append(handlers, route.Handler)...)
		}
	}

	return nil
}