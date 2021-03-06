package lastseen

import (
	"context"
)

// Exporter defines a last seen metrics exporter
type Exporter interface {
	Export(ctx context.Context, field string, val LastSeen) error
}

// ExporterFactory defines a exporter factory method
type ExporterFactory func() (Exporter, error)

// Multi returns an exporter that invokes all the exporters given to it in order.
func Multi(e ...Exporter) Exporter {
	a := make(multi, 0, len(e))
	for _, i := range e {
		if i == nil {
			continue
		}
		if i, ok := i.(multi); ok {
			a = append(a, i...)
			continue
		}
		a = append(a, i)
	}
	return a
}

type multi []Exporter

func (m multi) Export(ctx context.Context, field string, val LastSeen) error {
	for _, exporter := range m {
		err := exporter.Export(ctx, field, val)
		if err != nil {
			return err
		}
	}
	return nil
}
