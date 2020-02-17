package lastseen

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/wrapp/instrumentation/logs"
)

// Tracker takes care of tracking and exporting metrics
type Tracker interface {
	// SetSeen records a seen timestamp for the given field
	SetSeen(ctx context.Context, field string)
	// GetSeen returns the unix time stamp when the given field was last tracked as seen
	GetSeen(ctx context.Context, field string) int64
	// Wait for all exports to finish
	WaitForExporters()
}

// LastSeen keeps track of a last seen and its last flush time
type LastSeen struct {
	Value     int64
	LastFlush int64
}

// TrackerOption defines a configuration function
type TrackerOption func(*TrackerOptions)

// TrackerOptions defines options for the tracker
type TrackerOptions struct {
	flushIntervalSeconds int64
	useTicker            bool
	exporterFactories    []ExporterFactory
}

type tracker struct {
	seens                sync.Map
	exporterMu           sync.Mutex
	exporter             Exporter
	wg                   sync.WaitGroup
	flushIntervalSeconds int64
	useTicker            bool
}

var (
	once            sync.Once
	lastSeenTracker Tracker
)

// NewTrackerSingleton initializes a last seen tracker singleton
func NewTrackerSingleton(ctx context.Context, options ...TrackerOption) (Tracker, error) {

	opts := TrackerOptions{
		flushIntervalSeconds: 30,
		useTicker:            false,
	}
	for _, apply := range options {
		apply(&opts)
	}

	var exporters []Exporter
	for _, factory := range opts.exporterFactories {
		exporter, err := factory()
		if err != nil {
			return nil, err
		}
		exporters = append(exporters, exporter)
	}

	once.Do(func() {

		t := &tracker{
			useTicker:            opts.useTicker,
			flushIntervalSeconds: opts.flushIntervalSeconds,
			exporter:             Multi(exporters...),
		}

		if opts.useTicker {
			t.start()
		}

		lastSeenTracker = t
	})
	return lastSeenTracker, nil
}

// GetTracker gets the last seen tracker singleton
func GetTracker() Tracker {
	if lastSeenTracker == nil {
		return nil
	}
	return lastSeenTracker
}

// SetSeen updates the last seen time for the given field
func (t *tracker) SetSeen(ctx context.Context, field string) {
	if lastSeenTracker == nil {
		return
	}
	t.setLastSeen(ctx, field, time.Now().Unix())
}

// GetSeen gets the last seen time for the given field, if it has never been set the time is 0
func (t *tracker) GetSeen(ctx context.Context, field string) int64 {
	if lastSeenTracker == nil {
		return 0
	}

	if lastSeen, ok := t.getLastSeen(field); ok {
		return lastSeen.Value
	}
	return 0
}

// WaitForExporters waits for all exporters to finish the current exports
func (t *tracker) WaitForExporters() {
	if lastSeenTracker == nil {
		return
	}
	t.wg.Wait()
}

func (t *tracker) setLastSeen(ctx context.Context, field string, value int64) {
	var lastSeen LastSeen
	if existing, ok := t.getLastSeen(field); !ok {
		lastSeen = LastSeen{value, 0}
	} else if existing.Value == value {
		return // same value no need to update anything
	} else {
		lastSeen = LastSeen{value, existing.LastFlush}
	}
	t.seens.Store(field, lastSeen)
	// Flush directly unless we are using a timer
	if !t.useTicker {
		t.wg.Add(1)
		go func(trk *tracker) {
			defer trk.wg.Done()
			trk.flushKey(ctx, field, lastSeen)
		}(t)
	}
}

func (t *tracker) getLastSeen(field string) (LastSeen, bool) {
	if sv, ok := t.seens.Load(field); ok {
		return sv.(LastSeen), true
	}
	return LastSeen{}, false
}

func (t *tracker) start() {
	go func() {
		// run last-seen updater in background not to interfere with
		// async flow.
		for {
			<-time.After(time.Duration(t.flushIntervalSeconds) * time.Second)
			t.flushAll(context.Background())
		}
	}()
}

func (t *tracker) flushKey(ctx context.Context, field string, lastSeen LastSeen) {
	diff := lastSeen.Value - lastSeen.LastFlush
	if diff > 0 {
		if diff > t.flushIntervalSeconds {
			//logs.New(ctx).Info().Str("type", field).Int64("lastSeen", lastSeen.Value).Int64("lastFlush", lastSeen.LastFlush).Msg("Exporting last seen")
			// Store new flush time to prevent export of same value
			t.seens.Store(field, LastSeen{lastSeen.Value, lastSeen.Value})
			err := t.export(ctx, field, lastSeen)
			if err != nil {
				logs.New(ctx).Error().Err(err).Str("type", field).Msg("failed to flush last seen")
				// revert flushtime
				t.seens.Store(field, LastSeen{lastSeen.Value, lastSeen.LastFlush})
			}
		}
	}
}

func (t *tracker) flushAll(ctx context.Context) {
	t.seens.Range(func(key, val interface{}) bool {
		field, lastSeen := key.(string), val.(LastSeen)
		t.flushKey(ctx, field, lastSeen)
		return true
	})
}

func (t *tracker) export(ctx context.Context, field string, lastSeen LastSeen) error {
	t.exporterMu.Lock()
	defer t.exporterMu.Unlock()
	return t.exporter.Export(ctx, field, lastSeen)
}

// WithFlushIntervalSeconds specifies minimum flush interval
func WithFlushIntervalSeconds(t int64) TrackerOption {
	return func(o *TrackerOptions) {
		o.flushIntervalSeconds = t
	}
}

// WithFlushIntervalSecondsString specifies minimum flush interval with a string that will be parsed as integer
func WithFlushIntervalSecondsString(s string) TrackerOption {
	return func(o *TrackerOptions) {
		if s == "" {
			return
		}
		t, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			o.flushIntervalSeconds = t
		}
	}
}

// WithTicker controls whether a timer should do flushing, otherwise the flushing happens when values updated
func WithTicker(f bool) TrackerOption {
	return func(o *TrackerOptions) {
		o.useTicker = f
	}
}

// WithExporter provides an exporter factory, can be called multiple times to provide multiple exporters
func WithExporter(f ExporterFactory) TrackerOption {
	return func(o *TrackerOptions) {
		o.exporterFactories = append(o.exporterFactories, f)
	}
}
