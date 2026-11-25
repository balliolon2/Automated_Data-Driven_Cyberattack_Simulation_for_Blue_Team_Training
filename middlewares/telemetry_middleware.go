package middlewares

import (
	"net/http"
	"time"

	"cybersim/services"

	"github.com/gin-gonic/gin"
)

// TelemetryMiddleware records normalized HTTP route latency and status classes into
// the bounded in-memory NFRTelemetryStore, excluding CORS preflights and authenticated internal probes.
func TelemetryMiddleware(store *services.NFRTelemetryStore) gin.HandlerFunc {
	if store == nil {
		store = services.GlobalNFRTelemetry()
	}
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions || store.IsInternalProbeRequest(c.Request) {
			c.Next()
			return
		}

		store.IncrementInFlight()
		start := time.Now()
		defer store.DecrementInFlight()

		c.Next()

		durationMs := float64(time.Since(start).Microseconds()) / 1000.0
		matched := c.FullPath()
		if c.Request.Method == http.MethodPost && matched == "/api/admin/nfr/telemetry/reset" {
			return
		}
		routeLabel := "<unmatched>"
		if matched != "" {
			routeLabel = c.Request.Method + " " + matched
		}
		store.RecordHTTPRequest(routeLabel, c.Writer.Status(), durationMs)
	}
}
