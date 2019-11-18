package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/wrapp/instrumentation/requestid"
	"go.opencensus.io/plugin/ochttp"
)

var (
	// ErrTimeout is an error raised when a timeout occurs
	ErrTimeout = errors.New("Timeout")
)

// Client is an instrumented client.
type Client interface {
	Get(ctx context.Context, url string, funcs ...RequestOption) (Response, error)
	Post(ctx context.Context, url string, funcs ...RequestOption) (Response, error)
	Put(ctx context.Context, url string, funcs ...RequestOption) (Response, error)
	Delete(ctx context.Context, url string, funcs ...RequestOption) (Response, error)
}

type client struct {
	client         http.Client
	spanNameFormat string
	serviceName    string
}

// New creates a new instrumented client
func New(funcs ...func(*client) error) (Client, error) {
	cli := client{
		serviceName:    os.Getenv("SERVICE_NAME"),
		spanNameFormat: fmt.Sprintf("from %s", os.Getenv("SERVICE_NAME")),
	}
	for _, apply := range funcs {
		if err := apply(&cli); err != nil {
			return nil, err
		}
	}

	return cli, nil
}

// Request contains the parameters of the request.
type Request struct {
	url          string
	method       string
	body         io.Reader
	headers      map[string]string
	maxRetry     *uint
	timeout      time.Duration
	failManagers []FailManager
}

// RequestOption is a function that can be injected in the request.
type RequestOption func(*Request) error

// Response returns the request's response.
type Response struct {
	StatusCode int
	Body       cancelableBody
}

type cancelableBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b cancelableBody) Close() error {
	if err := b.ReadCloser.Close(); err != nil {
		return err
	}
	b.cancel()
	return nil
}

func (c client) try(ctx context.Context, request Request, cancelFunc context.CancelFunc) (Response, error) {
	req, err := http.NewRequest(request.method, request.url, request.body)
	if err != nil {
		return Response{}, err
	}

	for k, v := range request.headers {
		req.Header.Set(k, v)
	}

	req = req.WithContext(ctx)
	c.client.Transport = &ochttp.Transport{
		FormatSpanName: func(*http.Request) string { return c.spanNameFormat },
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return Response{}, err
	}

	for _, fm := range request.failManagers {
		if err := fm.Check(resp); err != nil {
			return Response{}, err
		}
	}

	return Response{Body: cancelableBody{
		resp.Body,
		cancelFunc,
	}, StatusCode: resp.StatusCode}, nil
}

func (c client) do(ctx context.Context, url, method string, funcs ...RequestOption) (Response, error) {
	req := Request{
		url:     url,
		method:  method,
		timeout: time.Second,
	}

	// Applying "battery-included" options.
	Header("X-Request-ID", requestid.Get(ctx))(&req)
	UserAgent(c.serviceName)(&req)

	// Applying the "on-demand" options
	for _, apply := range funcs {
		if err := apply(&req); err != nil {
			return Response{}, err
		}
	}

	type result struct {
		resp Response
		err  error
	}

	ch := make(chan result)

	cancelableCtx, cancel := context.WithTimeout(ctx, req.timeout)

	go func(tryCount uint) {
		for {
			resp, err := c.try(cancelableCtx, req, cancel)
			if err != nil && req.maxRetry != nil && tryCount < *req.maxRetry {
				tryCount++
				continue
			}

			ch <- result{resp: resp, err: err}
			return
		}
	}(1)

	for {
		select {
		case <-cancelableCtx.Done():
			return Response{}, ErrTimeout
		case res := <-ch:
			return res.resp, res.err
		}
	}
}

func (c client) Get(ctx context.Context, url string, funcs ...RequestOption) (Response, error) {
	return c.do(ctx, url, http.MethodGet, funcs...)
}

func (c client) Post(ctx context.Context, url string, funcs ...RequestOption) (Response, error) {
	return c.do(ctx, url, http.MethodPost, funcs...)
}

func (c client) Put(ctx context.Context, url string, funcs ...RequestOption) (Response, error) {
	return c.do(ctx, url, http.MethodPut, funcs...)
}

func (c client) Delete(ctx context.Context, url string, funcs ...RequestOption) (Response, error) {
	return c.do(ctx, url, http.MethodDelete, funcs...)
}
