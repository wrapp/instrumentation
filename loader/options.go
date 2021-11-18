package loader

import (
	"time"

	dataloader "github.com/graph-gophers/dataloader"
)

// NewOptions creates new loader options
func NewOptions(opts ...Option) []Option {
	o := []dataloader.Option{}
	// Apply options
	for _, opt := range opts {
		o = append(o, opt)
	}
	return o
}

// WithCache sets the cache to use on loaders otherwise a in memory cache will be used that lives through the http request
func WithCache(c dataloader.Cache) Option {
	return dataloader.WithCache(c)
}

// WithTracer allows tracing of calls to Load and LoadMany
func WithTracer(tracer dataloader.Tracer) Option {
	return dataloader.WithTracer(tracer)
}

// WithWait sets amount of time to wait before triggering batch, default 16 ms
func WithWait(d time.Duration) Option {
	return dataloader.WithWait(d)
}
