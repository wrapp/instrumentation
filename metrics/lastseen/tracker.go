package lastseen

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/wrapp/instrumentation/logs"
)

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
}

type tracker struct {
	seens      stateMap
	options    TrackerOptions
	exporterMu sync.Mutex
	exporter   Exporter
}

var (
	once            sync.Once
	trackerMu       sync.Mutex
	lastSeenTracker *tracker
)

// NewRedisExporter initializes a Redis exporter for last seen metrics
func NewRedisExporter(ctx context.Context, serviceName string, host string, storeKey string, options ...TrackerOption) error {
	once.Do(func() {
		if host != "" {
			opts := TrackerOptions{
				flushIntervalSeconds: 30,
				useTicker:            false,
			}
			for _, apply := range options {
				apply(&opts)
			}
			if storeKey == "" {
				storeKey = "eventbus-metrics"
			}
			redis, err := Redis(host, serviceName, storeKey)
			if err != nil {
				logs.New(ctx).Panic().Err(err).Msg("Unable create Redis last seen tracker")
			}
			newTrackerSingleton(ctx, opts, redis)
		}
	})
	return nil
}

func newTrackerSingleton(ctx context.Context, options TrackerOptions, e ...Exporter) {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	lastSeenTracker = &tracker{
		seens:    stateMap{state: make(map[string]LastSeen)},
		options:  options,
		exporter: Multi(e...),
	}

	if options.useTicker {
		lastSeenTracker.start()
	}
}

// SetSeen updates the last seen time for the given field
func SetSeen(ctx context.Context, field string) {
	if lastSeenTracker == nil {
		return
	}
	trackerMu.Lock()
	defer trackerMu.Unlock()
	lastSeenTracker.SetSeen(ctx, field)
}

// Tracker takes care of tracking and exporting metrics
type Tracker interface {
	// SetSeen records a seen timestamp for the given field
	SetSeen(field string)
	// GetSeen returns the unix time stamp when the given field was last tracked as seen
	GetSeen(field string) int64
	// Flushes all last seen values to store
	Flush()
}

type stateMap struct {
	sync.RWMutex
	state map[string]LastSeen
}

func (t *stateMap) SetNow(field string) {
	t.Set(field, time.Now().Unix())
}

func (t *stateMap) Set(field string, value int64) {
	t.Lock()
	if sv, ok := t.state[field]; !ok {
		t.state[field] = LastSeen{value, 0}
	} else {
		t.state[field] = LastSeen{value, sv.LastFlush}
	}
	t.Unlock()
}

func (t *stateMap) MarkFlushed(field string) {
	t.Lock()
	if sv, ok := t.state[field]; ok {
		t.state[field] = LastSeen{sv.Value, sv.Value}
	}
	t.Unlock()
}

func (t *stateMap) Get(field string) LastSeen {
	t.RLock()
	defer t.RUnlock()
	return t.state[field]
}

func (t *tracker) start() {
	go func() {
		// run last-seen updater in background not to interfere with
		// async flow.
		for {
			<-time.After(time.Duration(t.options.flushIntervalSeconds) * time.Second)
			t.flushAll(context.Background())
		}
	}()
}

// UpdateSeen sets the last time for the given field to the current unix time
func (t *tracker) SetSeen(ctx context.Context, field string) {
	t.seens.SetNow(field)
	go t.flushKey(ctx, field)
}

// Seen returns the last seen unix time for the given field.
func (t *tracker) GetSeen(field string) int64 {
	return t.seens.Get(field).Value
}

func (t *tracker) flushAll(ctx context.Context) {
	for txType := range t.seens.state {
		t.flushKey(ctx, txType)
	}
}

func (t *tracker) flushKey(ctx context.Context, field string) {
	sv := t.seens.Get(field)
	diff := sv.Value - sv.LastFlush
	if diff > 0 {
		if diff > t.options.flushIntervalSeconds {
			err := t.flush(ctx, field, sv)
			if err != nil {
				logs.New(ctx).Error().Err(err).Str("type", field).Msg("failed to flush last seen")
			} else {
				t.seens.MarkFlushed(field)
			}
		}
	}
}

func (t *tracker) flush(ctx context.Context, field string, val LastSeen) error {
	t.exporterMu.Lock()
	defer t.exporterMu.Unlock()
	return t.exporter.Flush(ctx, field, val)
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
