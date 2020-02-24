package lastseen

type lastSeenError string

func (err lastSeenError) Error() string { return string(err) }

var (
	// ErrNoTracker is raised when no tracker has been configured
	ErrNoTracker = lastSeenError("the collector endpoint is not valid")
)
