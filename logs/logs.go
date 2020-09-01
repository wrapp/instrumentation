package logs

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/wrapp/instrumentation/requestid"
	"github.com/wrapp/instrumentation/sessionid"
)

var logger zerolog.Logger

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimestampFieldName = "timestamp"
	zerolog.MessageFieldName = "msg"

	logger = zerolog.New(os.Stderr)
}

// New returns a zerolog Logger.
func New(ctx context.Context, funcs ...func(zerolog.Context) zerolog.Context) *zerolog.Logger {
	log := logger.With()
	log = log.Timestamp()

	funcs = append(funcs, WithServiceName())
	funcs = append(funcs, WithRequestID(ctx))
	funcs = append(funcs, WithSessionID(ctx))
	for _, apply := range funcs {
		log = apply(log)
	}
	l := log.Logger()
	return &l
}

// Info returns a logger with an info log level.
func Info(ctx context.Context, funcs ...func(zerolog.Context) zerolog.Context) *zerolog.Event {
	return New(ctx, funcs...).Info()
}

// Warn returns a logger with a warning log level.
func Warn(ctx context.Context, funcs ...func(zerolog.Context) zerolog.Context) *zerolog.Event {
	return New(ctx, funcs...).Warn()
}

// Error returns a logger with an error log level.
func Error(ctx context.Context, funcs ...func(zerolog.Context) zerolog.Context) *zerolog.Event {
	return New(ctx, funcs...).Error()
}

// Panic returns a logger with a panic log level.
func Panic(ctx context.Context, funcs ...func(zerolog.Context) zerolog.Context) *zerolog.Event {
	return New(ctx, funcs...).Panic()
}

// WithServiceName adds the service name to the logs.
func WithServiceName() func(zerolog.Context) zerolog.Context {
	return func(log zerolog.Context) zerolog.Context {
		serviceName := os.Getenv("SERVICE_NAME")
		if serviceName == "" {
			return log
		}
		return log.Str("service", serviceName)
	}
}

// WithRequestID adds the request id to the logs.
func WithRequestID(ctx context.Context) func(zerolog.Context) zerolog.Context {
	return func(log zerolog.Context) zerolog.Context {
		requestID := requestid.Get(ctx)
		if requestID == "" {
			return log
		}
		return log.Str("request_id", requestID)
	}
}

// WithSessionID injects the session id in the logs.
func WithSessionID(ctx context.Context) func(zerolog.Context) zerolog.Context {
	return func(log zerolog.Context) zerolog.Context {
		sessionID := sessionid.Get(ctx)
		if sessionID == "" {
			return log
		}
		return log.Str("session_id", sessionID)
	}
}

// MaskSSN masks the SSN from the logs.
func MaskSSN(ssn string) string {
	if len(ssn) < 4 {
		// seems unlikely to have a 4 digits SSN but in any case, we obscure it.
		return "****"
	}

	mask := strings.Repeat("*", len(ssn)-4)
	suffix := ssn[len(ssn)-4:]

	return mask + suffix
}
