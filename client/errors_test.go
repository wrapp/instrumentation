package client_test

import (
	"testing"

	"github.com/wrapp/instrumentation/client"
)

const validationErrorResponse = `{
	"valid": false,
	"errors": [
	  {
		"context": "(root).address",
		"expected": "array",
		"field": "address",
		"given": "null"
	  },
	  {
		"context": "(root).postal_code",
		"expected": "string",
		"field": "postal_code",
		"given": "null"
	  },
	  {
		"context": "(root).address_line2",
		"expected": "string",
		"field": "address_line2",
		"given": "null"
	  },
	  {
		"context": "(root).city",
		"expected": "string",
		"field": "city",
		"given": "null"
	  },
	  {
		"context": "(root).country_code",
		"expected": "string",
		"field": "country_code",
		"given": "null"
	  }
	],
	"item": 0
  }`

func TestParseValidationErrorMessageOveride(t *testing.T) {
	ve, err := client.ParseValidationErrors([]byte(validationErrorResponse), "Test")
	if err != nil {
		t.Error(err)
	}

	// Providing custom message overrides the default message
	if ve.Error() != "Test" {
		t.Fail()
	}
}

func TestParseValidationErrorWithoutMessageOverride(t *testing.T) {
	ve, err := client.ParseValidationErrors([]byte(validationErrorResponse), "")
	if err != nil {
		t.Error(err)
	}

	// Providing custom message overrides the default message
	if ve.Error() == "" {
		t.Fail()
	}
}
