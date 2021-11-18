package loader

import (
	"context"
	"fmt"
	"time"

	dataloader "github.com/graph-gophers/dataloader"
)

type key string

var loadersKey key = "dataloaders"

type loaderDescription struct {
	loadFunc     LoadFunc
	useAppCache  bool
	useAppTracer bool
	wait         time.Duration
}

var registeredLoaders map[key]loaderDescription

func init() {
	registeredLoaders = map[key]loaderDescription{}
}

// RegisterLoader registers a loader to the registry of loaders to initialize
func RegisterLoader(loaderKey key, l LoadFunc, useAppCache bool, useAppTracer bool, wait time.Duration) {
	registeredLoaders[loaderKey] = loaderDescription{
		loadFunc:     l,
		useAppCache:  useAppCache,
		useAppTracer: useAppTracer,
		wait:         wait,
	}
}

// Loaders has all the available data loaders
type Loaders struct {
	loaderMap map[key]Loader
}

// NewLoader creats new loader from a loader function and options
func NewLoader(f LoadFunc, opts ...Option) *Loader {
	return dataloader.NewBatchedLoader(f, opts...)
}

// CreateLoaders
func CreateLoaders(appCache Cache, appTracer Tracer) *Loaders {
	l := &Loaders{loaderMap: make(map[key]dataloader.Loader)}

	for k, reg := range registeredLoaders {
		opts := []Option{}
		if reg.useAppCache {
			opts = append(opts, WithCache(appCache))
		}
		if reg.useAppTracer {
			opts = append(opts, WithTracer(appTracer))
		}
		if reg.wait > 0 {
			opts = append(opts, WithWait(reg.wait))
		}
		l.loaderMap[k] = *NewLoader(reg.loadFunc, opts...)
	}
	return l
}

// Use picks the loader to use with a loader key.
func (l *Loaders) Use(loaderKey key) *Loader {
	if loader, ok := l.loaderMap[loaderKey]; ok {
		return &loader
	}
	return dataloader.NewBatchedLoader(func(_ context.Context, keys Keys) []*Result {
		var results []*Result
		for _, k := range keys {
			results = append(results, &Result{Data: k, Error: fmt.Errorf("unregistered data loader %s", string(loaderKey))})
		}
		return results
	})
}
