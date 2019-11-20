package requestid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wrapp/instrumentation/client"
	"github.com/wrapp/instrumentation/requestid"
)

func TestStoreAndGet(t *testing.T) {
	expectedRequestID := "some-request-id"
	ctx := requestid.Store(context.Background(), expectedRequestID)

	requestID := requestid.Get(ctx)
	if requestID != expectedRequestID {
		t.Fatalf("expected %v got %v", expectedRequestID, requestID)
	}
}

func TestEmptyRequestID(t *testing.T) {
	requestID := requestid.Get(context.Background())
	if requestID != "" {
		t.Fatalf("expected empty string got %v", requestID)
	}
}

func TestMiddleware(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(requestid.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestID := requestid.Get(r.Context())
			if requestID == "" {
				t.Fatalf("expected a non-empty string, got %v", requestID)
			}
		})))
	cli, _ := client.New()
	_, err := cli.Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("got an unexpected error %v", err)
	}
}
