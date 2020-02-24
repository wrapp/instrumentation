package jsonvalidation_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wrapp/instrumentation/client"
	"github.com/wrapp/instrumentation/jsonvalidation"
)

func TestMiddleware(t *testing.T) {
	ctx := context.Background()

	schema := strings.NewReader(`{
		"$schema": "http://json-schema.org/schema#",
		"type": "object",
		"properties": {
			"national_id": {
				"type": "string"
			},
			"country": {
				"type": "string"
			}
		},
		"required": ["national_id", "country"],
		"additionalProperties": false
	}`)
	server := httptest.NewServer(
		jsonvalidation.Middleware(schema)(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {

			})))
	error400 := errors.New("http400")

	tests := []struct {
		testcase      string
		payload       []byte
		expectedError error
	}{
		{
			testcase: "should not return an error when the payload is valid",
			payload:  []byte(`{"national_id": "123456780000", "country": "SE"}`),
		},
		{
			testcase:      "should return an error when the payload is empty",
			expectedError: error400,
		},
		{
			testcase:      "should return an error when the payload is invalid",
			payload:       []byte(`{"national_id": 3.14159265359, "country": "SE"}`),
			expectedError: error400,
		},
		{
			testcase:      "should return an error when the payload is missing some field",
			payload:       []byte(`{"country": "SE"}`),
			expectedError: error400,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.testcase, func(t *testing.T) {
			cli, _ := client.New()
			_, err := cli.Post(ctx, server.URL,
				client.FailOn(client.StatusChecker(error400,
					http.StatusBadRequest)),
				client.Body(test.payload),
			)
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("got an unexpected error %v", err)
			}
		})
	}
}
