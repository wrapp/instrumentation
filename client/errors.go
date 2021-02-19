package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	fieldErrMsg = "Context: '%s' Error:Field validation for '%s' failed. Expected: %s but given: %s"
)

// FieldError describes error of a field
type FieldError struct {
	Context  string `json:"context"`
	Field    string `json:"field"`
	Expected string `json:"expected"`
	Given    string `json:"given"`
}

func (fe FieldError) Error() string {
	return fmt.Sprintf(fieldErrMsg, fe.Context, fe.Field, fe.Expected, fe.Given)
}

// ValidationErrors is an array of FieldError's
// for use in custom error messages post validation.
type ValidationErrors struct {
	Valid  bool         `json:"valid"`
	Errors []FieldError `json:"errors"`
}

// Error creates an error message from array of validation errors
func (ve ValidationErrors) Error() string {

	buff := bytes.NewBufferString("Validation Errors:\n")

	var fe *FieldError

	for i := 0; i < len(ve.Errors); i++ {

		fe = &ve.Errors[i]
		buff.WriteString(fe.Error())
		buff.WriteString("\n")
	}

	return strings.TrimSpace(buff.String())
}

// Extensions provide a map of the error properties
func (ve ValidationErrors) Extensions() map[string]interface{} {
	m := map[string]interface{}{
		"message": "Field validation",
		"valid":   ve.Valid,
		"fields":  ve.Errors,
	}
	return m
}

// ParseValidationErrors parses a byte array for validation errors
func ParseValidationErrors(body []byte) (ValidationErrors, error) {
	var validationErrors ValidationErrors
	if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&validationErrors); err != nil {
		return ValidationErrors{}, fmt.Errorf("Failed to parse validation errors: %w", err)
	}
	return validationErrors, nil
}
