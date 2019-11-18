package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/wrapp/instrumentation/client"
	"github.com/wrapp/instrumentation/requestid"
)

type payload struct {
	ID  string `json:"id"`
	Msg string `json:"msg"`
}

func TestSimple(t *testing.T) {
	ctx := context.Background()

	expected := payload{ID: "42", Msg: "the answer to life the universe and everything"}
	expectedUserAgent := "ua-test"

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.UserAgent() != "ua-test" {
				t.Fatalf("expected user-agent %s, got %s", expectedUserAgent,
					r.UserAgent())
			}
			requestID := r.Header.Get("X-Request-ID")
			if requestID != "request-id" {
				t.Fatalf("expected request-id 'request-id', got %s", requestID)
			}
			json.NewEncoder(w).Encode(expected)
		}))
	ctx = requestid.Store(ctx, "request-id")
	os.Setenv("SERVICE_NAME", expectedUserAgent)

	cli, _ := client.New()
	resp, err := cli.Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("expected no errors got %v", err)
	}
	defer resp.Body.Close()

	var got payload
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unable to read body, got %v", err)
	}

	if got != expected {
		t.Fatalf("expected %v got %v", expected, got)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestBody(t *testing.T) {
	ctx := context.Background()
	buffer, _ := json.Marshal(`{"id": 1337, "msg": "yo"}`)

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			b, _ := ioutil.ReadAll(r.Body)
			if !bytes.Equal(b, buffer) {
				t.Fatalf("expected %s got %s", string(buffer), string(b))
			}
			w.WriteHeader(http.StatusNoContent)
		}))

	cli, _ := client.New()

	resp, err := cli.Post(ctx, server.URL, client.Body(buffer))
	if err != nil {
		t.Fatalf("expected no errors got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestFailOnStatus(t *testing.T) {
	ctx := context.Background()

	error400 := errors.New("http400")
	error500 := errors.New("http500")

	tests := []struct {
		testcase      string
		statusCode    int
		expectedError error
	}{
		{
			testcase:   "should not fail",
			statusCode: http.StatusOK,
		},
		{
			testcase:      "should fail with a 500 error",
			statusCode:    http.StatusInternalServerError,
			expectedError: error500,
		},
		{
			testcase:      "should fail with a 400 error",
			statusCode:    http.StatusBadRequest,
			expectedError: error400,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.testcase, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(test.statusCode)
				}))

			cli, _ := client.New()
			_, err := cli.Get(ctx, server.URL,
				client.FailOn(client.StatusChecker(error500, http.StatusInternalServerError)),
				client.FailOn(client.StatusChecker(error400, http.StatusBadRequest)))
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expected %v got %v", test.expectedError, err)
			}
		})
	}
}

func TestTimeout(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			<-time.After(50 * time.Millisecond)
			w.WriteHeader(http.StatusInternalServerError)
		}))

	cli, _ := client.New()
	_, err := cli.Put(ctx, server.URL, client.Timeout(10*time.Millisecond))
	if !errors.Is(err, client.ErrTimeout) {
		t.Fatalf("expected %v got %v", client.ErrTimeout, err)
	}
}

func TestRetries(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		testcase        string
		expectedCounter uint
		retries         uint
	}{
		{
			testcase:        "should query only once if no retries are set",
			retries:         uint(0),
			expectedCounter: uint(1),
		},
		{
			testcase:        "should query only once if the retry is set to 1",
			retries:         uint(1),
			expectedCounter: uint(1),
		},
		{
			testcase:        "should query 3 time if the retry is set to 3",
			retries:         uint(3),
			expectedCounter: uint(3),
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.testcase, func(t *testing.T) {
			counter := uint(0)
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					counter++
					w.WriteHeader(http.StatusInternalServerError)
				}))

			cli, _ := client.New()
			cli.Delete(ctx, server.URL,
				client.FailOn(client.StatusChecker(errors.New("oops"), http.StatusInternalServerError)),
				client.Retry(test.retries),
			)
			if counter != test.expectedCounter {
				t.Fatalf("expected %d got %d", test.expectedCounter, counter)
			}
		})
	}
}

func TestCancelable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			<-time.After(50 * time.Millisecond)
			w.WriteHeader(http.StatusInternalServerError)
		}))

	ch := make(chan error)
	go func() {
		cli, _ := client.New()
		_, err := cli.Put(ctx, server.URL)
		ch <- err
	}()

	cancel()
	err := <-ch
	if !errors.Is(err, client.ErrTimeout) {
		t.Fatalf("expected %v got %v\n", client.ErrTimeout, err)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	ctx := context.Background()

	type test struct {
		testcase            string
		backoff             time.Duration
		expectedRetries     uint
		expectedMinDuration time.Duration
	}

	tests := []test{
		test{
			testcase:        "no backoff should be really quick",
			backoff:         0 * time.Millisecond,
			expectedRetries: uint(3),
		},
		test{
			testcase:            "50ms backoff with 3 retries should be at least 200ms",
			backoff:             50 * time.Millisecond,
			expectedRetries:     uint(3),
			expectedMinDuration: 200 * time.Millisecond,
		},
		test{
			testcase:            "50ms backoff with 4 retries should be at least 550ms",
			backoff:             50 * time.Millisecond,
			expectedRetries:     uint(4),
			expectedMinDuration: 550 * time.Millisecond,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.testcase, func(t *testing.T) {
			counter := uint(0)
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					counter++
					w.WriteHeader(http.StatusInternalServerError)
				}))

			start := time.Now().UTC()
			cli, _ := client.New()
			cli.Post(ctx, server.URL,
				client.FailOn(client.StatusChecker(errors.New("oops"),
					http.StatusInternalServerError)),
				client.RetryWithBackoff(test.expectedRetries, test.backoff),
			)
			if counter != test.expectedRetries {
				t.Fatalf("expected %d got %d", test.expectedRetries, counter)
			}

			duration := time.Since(start)
			if duration < test.expectedMinDuration {
				t.Fatalf("expected %v got %v", test.expectedMinDuration, duration)
			}
		})
	}
}
