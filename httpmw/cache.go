package httpmw

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type cacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *cacheWriter) WriteString(s string) (int, error) {
	w.body.Write([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

// Cache middleware to cache http responses
func Cache(cache redis.Cmdable, expire time.Duration, handle gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL.RequestURI()
		data, err := cache.Get(c, url).Bytes()

		if err == nil {
			c.Data(http.StatusOK, http.DetectContentType(data), data)
			return
		}

		if err != redis.Nil {
			log.Println(err)
		}

		cw := new(cacheWriter)
		cw.body = bytes.NewBuffer([]byte{})
		cw.ResponseWriter = c.Writer
		c.Writer = cw
		handle(c)

		if cw.Status() != http.StatusOK {
			return
		}

		if err := cache.Set(c, url, cw.body.Bytes(), expire).Err(); err != nil {
			log.Println(err)
		}
	}
}
