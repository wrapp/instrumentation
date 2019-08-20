package logs

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/wrapp/instrumentation/requestid"
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

	funcs = append(funcs, WithServiceName())
	funcs = append(funcs, WithRequestID(ctx))
	for _, apply := range funcs {
		log = apply(log)
	}
	l := log.Logger()
	return &l
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
