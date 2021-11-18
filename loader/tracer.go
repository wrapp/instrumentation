package loader

import (
	"context"
	"strings"

	dataloader "github.com/graph-gophers/dataloader"
	"github.com/wrapp/instrumentation/tracing"
)

// LogTracer Tracer implements a tracer using the tracing instrumentation
type LogTracer struct{}

// TraceLoad will trace a call to dataloader.LoadMany
func (LogTracer) TraceLoad(ctx context.Context, key dataloader.Key) (context.Context, dataloader.TraceLoadFinishFunc) {
	ctx, callback := tracing.Span(ctx, tracing.Label("Dataloader.load"), tracing.StringTags("dataloader.key", key.String()))
	defer callback()
	return ctx, func(thunk dataloader.Thunk) {}
}

// TraceLoadMany will trace a call to dataloader.LoadMany
func (LogTracer) TraceLoadMany(ctx context.Context, keys dataloader.Keys) (context.Context, dataloader.TraceLoadManyFinishFunc) {
	ctx, callback := tracing.Span(ctx, tracing.Label("Dataloader.loadmany"),
		tracing.StringTags("dataloader.keys", strings.Join(keys.Keys(), ",")))
	defer callback()
	return ctx, func(thunk dataloader.ThunkMany) {}
}

// TraceBatch will trace a call to dataloader.LoadMany
func (LogTracer) TraceBatch(ctx context.Context, keys dataloader.Keys) (context.Context, dataloader.TraceBatchFinishFunc) {
	ctx, callback := tracing.Span(ctx, tracing.Label("Dataloader.batch"),
		tracing.StringTags("dataloader.keys", strings.Join(keys.Keys(), ",")))
	defer callback()
	return ctx, func(results []*dataloader.Result) {}
}
