package tracing

import (
	"fmt"
	"net/http"

	"go.opencensus.io/plugin/ochttp"
)

// Transport traces outgoing request.
func Transport(serviceName string) *ochttp.Transport {
	return &ochttp.Transport{
		FormatSpanName: func(*http.Request) string { return fmt.Sprintf("from %s", serviceName) },
	}
}
