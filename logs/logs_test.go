package logs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/wrapp/instrumentation/logs"
	"github.com/wrapp/instrumentation/requestid"
	"github.com/wrapp/instrumentation/sessionid"
)

func TestMaskSSN(t *testing.T) {
	tests := []struct {
		testcase          string
		ssn               string
		expectedMaskedSSN string
	}{
		{
			testcase:          "should mask yyyymmdd-1234 to *********1234",
			ssn:               "yyyymmdd-1234",
			expectedMaskedSSN: "*********1234",
		},
		{
			testcase:          "should mask 123 to ****",
			ssn:               "123",
			expectedMaskedSSN: "****",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.testcase, func(t *testing.T) {
			masked := logs.MaskSSN(test.ssn)
			if masked != test.expectedMaskedSSN {
				t.Fatalf("expected %v got %v", test.expectedMaskedSSN, masked)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type log struct {
		Level     string `json:"level"`
		RequestID string `json:"request_id"`
		SessionID string `json:"session_id"`
		Service   string `json:"service"`
		Msg       string `json:"msg"`
	}

	expected := log{
		Level:     "info",
		RequestID: "my-request-id",
		SessionID: "my-session-id",
		Service:   "my-service-name",
		Msg:       "my-message",
	}

	os.Setenv("SERVICE_NAME", expected.Service)
	ctx := requestid.Store(context.Background(), expected.RequestID)
	ctx = sessionid.Store(ctx, expected.SessionID)

	var out bytes.Buffer
	logger := logs.New(ctx, logs.WithRequestID(ctx)).Output(&out)

	logger.Info().Msg(expected.Msg)

	var got log
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("got an unexpected error %v", err)
	}

	if got != expected {
		t.Fatalf("expected %v got %v", expected, got)
	}
}
