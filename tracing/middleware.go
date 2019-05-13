package tracing

import (
	"fmt"
	"net/http"
	"strings"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
)

// Middleware is a http middleware that intercepts the query and traces it.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.UserAgent()

		// we don't want to trace anything from the HealthChecker
		if strings.Contains(userAgent, "HealthChecker") {
			next.ServeHTTP(w, r)
			return
		}

		instrumentedHandler := ochttp.Handler{
			Handler:     next,
			Propagation: &b3.HTTPFormat{},
		}
		instrumentedHandler.FormatSpanName = func(r *http.Request) string {
			return fmt.Sprintf("from %s", userAgent)
		}
		instrumentedHandler.ServeHTTP(w, r)
	})
}
