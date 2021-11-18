package loader

import (
	dataloader "github.com/graph-gophers/dataloader"
)

// Key is the interface that all keys need to implement
type Key = dataloader.Key

// Keys wraps a slice of Key types to provide some convenience methods.
type Keys = dataloader.Keys

// Result is the data structure that a BatchFunc returns.
// It contains the resolved data, and any errors that may have occurred while fetching the data.
type Result = dataloader.Result

// ResultMany is used by the LoadMany method.
// It contains a list of resolved data and a list of errors.
// The lengths of the data list and error list will match, and elements at each index correspond to each other.
type ResultMany = dataloader.ResultMany

// Loader implements the dataloader.Interface.
type Loader = dataloader.Loader

// LoadFunc is a function, which when given a slice of keys (string), returns an slice of `results`.
// It's important that the length of the input keys matches the length of the output results.
//
// The keys passed to this function are guaranteed to be unique
type LoadFunc = dataloader.BatchFunc

// The Cache interface. If a custom cache is provided, it must implement this interface.
type Cache = dataloader.Cache

// Tracer is an interface that may be used to implement tracing.
type Tracer = dataloader.Tracer

// Option allows for configuration of Loader fields.
type Option = dataloader.Option
