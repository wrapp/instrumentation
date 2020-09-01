package sessionid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wrapp/instrumentation/client"
	"github.com/wrapp/instrumentation/requestid"
	"github.com/wrapp/instrumentation/sessionid"
)

func TestStoreAndGet(t *testing.T) {
	expectedSessionID := "some-request-id"
	ctx := requestid.Store(context.Background(), expectedSessionID)

	requestID := requestid.Get(ctx)
	if requestID != expectedSessionID {
		t.Fatalf("expected %v got %v", expectedSessionID, requestID)
	}
}
func TestMiddleware(t *testing.T) {
	const expected = "some-session-id"

	server := httptest.NewServer(sessionid.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			sessionID := sessionid.Get(r.Context())
			if sessionID != expected {
				t.Fatalf("expected a %s, got %v", expected, sessionID)
			}
		})))
	cli, _ := client.New()
	_, err := cli.Get(context.Background(), server.URL,
		client.Header("X-Session-ID", expected))
	if err != nil {
		t.Fatalf("got an unexpected error %v", err)
	}
}
