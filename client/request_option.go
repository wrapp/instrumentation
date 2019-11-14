package client

import "bytes"

import "time"

// Header adds a header to the request.
func Header(key string, value string) RequestOption {
	return func(req *Request) error {
		if req.headers == nil {
			req.headers = make(map[string]string)
		}

		req.headers[key] = value

		return nil
	}
}

// Body adds a payload to the request.
func Body(buffer []byte) RequestOption {
	return func(req *Request) error {
		req.body = bytes.NewBuffer(buffer)
		return nil
	}
}

// UserAgent adds the user-agent to the request.
func UserAgent(ua string) RequestOption {
	return Header("User-Agent", ua)
}

// FailOn adds a failure manager to the request.
func FailOn(fm FailManager) RequestOption {
	return func(req *Request) error {
		req.failManagers = append(req.failManagers, fm)

		return nil
	}
}

// Timeout adds a timeout to the request
func Timeout(duration time.Duration) RequestOption {
	return func(req *Request) error {
		req.timeout = duration
		return nil
	}
}

// Retry allows to retry the request multiple time.
func Retry(count uint) RequestOption {
	return func(req *Request) error {
		req.maxRetry = &count
		return nil
	}
}
