package httphandler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StatusResponse health check response
type StatusResponse struct {
	Uptime int               `json:"uptime"`
	Online map[string]bool   `json:"online"`
	Errors map[string]string `json:"errors"`
}

// StatusCheck check status of the service
type StatusCheck func(ctx context.Context) error

// Status health check API endpoint
func Status(services map[string]StatusCheck) gin.HandlerFunc {
	startup := time.Now().UTC()

	return func(c *gin.Context) {
		res := StatusResponse{
			Uptime: int(time.Since(startup).Seconds()),
			Errors: map[string]string{},
			Online: map[string]bool{},
		}

		for name, ping := range services {
			err := ping(c)

			if err != nil {
				res.Errors[name] = err.Error()
			}

			res.Online[name] = err == nil
		}

		c.JSON(http.StatusOK, res)
	}
}
