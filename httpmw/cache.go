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

type CacheParams struct {
	Cache       redis.Cmdable
	Expire      time.Duration
	Handle      gin.HandlerFunc
	ContentType string
}

// Cache middleware to cache http responses
func Cache(p *CacheParams) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL.RequestURI()
		data, err := p.Cache.Get(c, url).Bytes()

		if err == nil {
			c.Data(http.StatusOK, p.ContentType, data)
			return
		}

		if err != redis.Nil {
			log.Println(err)
		}

		cw := new(cacheWriter)
		cw.body = bytes.NewBuffer([]byte{})
		cw.ResponseWriter = c.Writer
		c.Writer = cw
		p.Handle(c)

		if cw.Status() != http.StatusOK {
			return
		}

		if err := p.Cache.Set(c, url, cw.body.Bytes(), p.Expire).Err(); err != nil {
			log.Println(err)
		}
	}
}
