package requestid

import "context"

type key string

const (
	contextRequestID key = "request-id"
)

// Store stores the request ID in the given context.
func Store(parent context.Context, requestID string) context.Context {
	return context.WithValue(parent, contextRequestID, requestID)
}

// Get returns the request ID from the context.
func Get(ctx context.Context) string {
	return ctx.Value(contextRequestID).(string)
}
