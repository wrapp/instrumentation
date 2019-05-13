package tracing

type tracingError string

func (err tracingError) Error() string { return string(err) }

var (
	// ErrInvalidCollectorEndpoint is raised when the collector endpoint is not
	// valid (eg. empty).
	ErrInvalidCollectorEndpoint = tracingError("the collector endpoint is not valid")
	// ErrInvalidServiceName is raised when the service name is not valid (eg. empty).
	ErrInvalidServiceName = tracingError("the service name is not valid")
	// ErrUnableToSetupJaegerExporter is raised when an error occurs while setting up
	// the jaeger exporter.
	ErrUnableToSetupJaegerExporter = tracingError("unable to setup the jaeger exporter")
)
