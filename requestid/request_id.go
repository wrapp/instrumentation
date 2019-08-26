package requestid

import (
	"context"
	"net/http"

	"github.com/m4rw3r/uuid"
)

type key string

const (
	contextRequestID key = "request-id"
)

// Middleware injects a request-id in the context
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// No request id exists in the request or the context: Generate one!
			id, _ := uuid.V4()
			requestID = id.String()
		}

		ctx := Store(reqCtx, requestID)
		w.Header().Add("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Store stores the request ID in the given context.
func Store(parent context.Context, requestID string) context.Context {
	return context.WithValue(parent, contextRequestID, requestID)
}

// Get returns the request ID from the context.
func Get(ctx context.Context) string {
	v := ctx.Value(contextRequestID)
	if v == nil {
		return ""
	}

	requestID, ok := ctx.Value(contextRequestID).(string)
	if !ok {
		return ""
	}
	return requestID
}
