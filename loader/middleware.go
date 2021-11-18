package loader

import (
	"context"
	"net/http"
)

var loadersContextKey key = "dataloaders"

// Middleware is a http request middleware for adding data loaders to the context
func Middleware(loaders *Loaders) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, loadersContextKey, loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// For gets data loaders from the context
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersContextKey).(*Loaders)
}
