package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logging logs each HTTP request with status, latency, and request ID.
func Logging(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		evt := log.Info()
		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			evt = log.Error()
		}
		evt.Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Str("client_ip", c.ClientIP()).
			Dur("duration_ms", time.Since(start)).
			Msg("http request")
	}
}
