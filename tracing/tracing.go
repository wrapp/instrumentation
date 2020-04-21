package tracing

import (
	"context"
	"fmt"
	"sync"

	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/wrapp/instrumentation/requestid"
	"go.opencensus.io/trace"
)

var (
	once           sync.Once
	tracingEnabled bool
)

// JaegerOptions are the options passed to the NewJaeger method to initialize
// everything.
type JaegerOptions struct {
	collectorEndpoint string
	serviceName       string
	sampler           func() trace.Sampler
}

// NewJaegerExporterSingleton initializes a Jaeger exporter only once.
func NewJaegerExporterSingleton(collectorEndpoint, serviceName string,
	funcs ...func(*JaegerOptions)) error {

	if collectorEndpoint == "" {
		return ErrInvalidCollectorEndpoint
	}

	if serviceName == "" {
		return ErrInvalidServiceName
	}

	var err error
	once.Do(func() {
		options := JaegerOptions{
			collectorEndpoint: collectorEndpoint,
			serviceName:       serviceName,
			sampler:           trace.AlwaysSample,
		}
		for _, apply := range funcs {
			apply(&options)
		}

		exporter, err := jaeger.NewExporter(jaeger.Options{
			CollectorEndpoint: options.collectorEndpoint,
			ServiceName:       options.serviceName,
		})
		if err != nil {
			err = ErrUnableToSetupJaegerExporter
			return
		}

		trace.ApplyConfig(trace.Config{
			DefaultSampler: options.sampler(),
		})
		trace.RegisterExporter(exporter)

		tracingEnabled = true
	})

	return err
}

// SpanOptions are a list of fields to include in the span.
type SpanOptions struct {
	Namespace  string
	Int64Tags  map[string]int64
	StringTags map[string]string
	label      string
}

func (o SpanOptions) spanLabel() string {
	if o.Namespace == "" {
		return o.label
	}

	return fmt.Sprintf("%s::%s", o.Namespace, o.label)
}

// StartSpan creates a new span
func StartSpan(ctx context.Context, label string, funcs ...func(*SpanOptions)) (context.Context, func()) {
	if !tracingEnabled {
		return ctx, func() {}
	}

	options := SpanOptions{label: label}
	for _, apply := range funcs {
		apply(&options)
	}

	ctx, span := trace.StartSpan(ctx, options.spanLabel())

	attributes := make([]trace.Attribute, 0, len(options.Int64Tags)+
		len(options.StringTags)+1)
	for k, v := range options.Int64Tags {
		attributes = append(attributes, trace.Int64Attribute(k, v))
	}

	for k, v := range options.StringTags {
		attributes = append(attributes, trace.StringAttribute(k, v))
	}

	attributes = append(attributes, trace.StringAttribute("request_id",
		requestid.Get(ctx)))

	span.AddAttributes(attributes...)

	return ctx, func() {
		span.End()
	}
}

// Namespace adds a name space to the span.
func Namespace(ns string) func(*SpanOptions) {
	return func(o *SpanOptions) {
		o.Namespace = ns
	}
}

// StringTags adds a tag to the span.
func StringTags(key, value string) func(*SpanOptions) {
	return func(o *SpanOptions) {
		if o.StringTags == nil {
			o.StringTags = make(map[string]string)
		}

		o.StringTags[key] = value
	}
}

// Int64Tags adds a int64 tag to the span.
func Int64Tags(key string, value int64) func(*SpanOptions) {
	return func(o *SpanOptions) {
		if o.Int64Tags == nil {
			o.Int64Tags = make(map[string]int64)
		}

		o.Int64Tags[key] = value
	}
}
