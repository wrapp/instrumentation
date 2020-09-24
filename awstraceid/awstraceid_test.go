package awstraceid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wrapp/instrumentation/awstraceid"
	"github.com/wrapp/instrumentation/client"
)

func TestStoreAndGet(t *testing.T) {
	expectedAwsTraceID := "some-banan-id"
	ctx := awstraceid.Store(context.Background(), expectedAwsTraceID)
	gotTraceID := awstraceid.Get(ctx)
	if gotTraceID != expectedAwsTraceID {
		t.Fatalf("expected %v got %v", expectedAwsTraceID, gotTraceID)
	}
}
func TestMiddleware(t *testing.T) {
	const expected = "some-banan-id"

	server := httptest.NewServer(awstraceid.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotTraceID := awstraceid.Get(r.Context())
			if gotTraceID != expected {
				t.Fatalf("expected a %s, got %v", expected, gotTraceID)
			}
		})))
	cli, _ := client.New()
	_, err := cli.Get(context.Background(), server.URL,
		client.Header("X-Amzn-Trace-Id", expected))
	if err != nil {
		t.Fatalf("got an unexpected error %v", err)
	}
}
