package sessionid

import (
	"context"
	"net/http"
)

type key string

const (
	contextSessionID key = "session-id"
)

// Middleware injects the session in the context
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("X-Session-ID", sessionID)
		next.ServeHTTP(w, r.WithContext(Store(r.Context(), sessionID)))
	})
}

// Store stores the request ID in the given context.
func Store(parent context.Context, sessionID string) context.Context {
	return context.WithValue(parent, contextSessionID, sessionID)
}

// Get returns the request ID from the context.
func Get(ctx context.Context) string {
	v := ctx.Value(contextSessionID)
	if v == nil {
		return ""
	}

	sessionID, ok := ctx.Value(contextSessionID).(string)
	if !ok {
		return ""
	}
	return sessionID
}
