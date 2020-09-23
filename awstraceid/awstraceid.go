package awstraceid

import (
	"context"
	"net/http"
)

type key string

const (
	// awsTraceIDKey used for the context-key
	awsTraceIDKey key = "amazonTraceId"
	// AWSTraceIDHeader is the httpHeader used for the tracing request through
	// aws load-balancers
	AWSTraceIDHeader = "X-Amzn-Trace-Id"
)

// Middleware injects the amazon-traceID in the context
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		awsTraceID := r.Header.Get(AWSTraceIDHeader)
		if awsTraceID == "" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add(AWSTraceIDHeader, awsTraceID)
		next.ServeHTTP(w, r.WithContext(Store(r.Context(), awsTraceID)))
	})
}

// Store returns a copy of the input context with the awsTraceId ammended.
func Store(parent context.Context, awsTraceID string) context.Context {
	return context.WithValue(parent, awsTraceIDKey, awsTraceID)
}

// Get returns the awsTraceID from the context.
func Get(ctx context.Context) string {
	v := ctx.Value(awsTraceIDKey)
	if v == nil {
		return ""
	}

	awsTraceID, ok := v.(string)
	if !ok {
		return ""
	}
	return awsTraceID
}
