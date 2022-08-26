package jsonvalidation

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/wrapp/instrumentation/logs"
	"github.com/xeipuuv/gojsonschema"
)

func badRequest(w http.ResponseWriter, r *http.Request, message, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	type response struct {
		Message string `json:"message,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	_ = json.NewEncoder(w).Encode(response{
		Message: message,
		Error:   err,
	})
}

// Middleware checks whether the request payload validates a given schema.
func Middleware(r io.Reader) func(next http.Handler) http.Handler {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		logs.New(context.Background()).
			Panic().
			Err(err).
			Msg("unable to load schema")
	}

	schemaLoader := gojsonschema.NewBytesLoader(b)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		logs.New(context.Background()).
			Panic().
			Err(err).
			Msg("unable to load schema")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buffer, err := ioutil.ReadAll(r.Body)
			if err != nil {
				badRequest(w, r, "no payload", err.Error())
				return
			}

			payload := gojsonschema.NewBytesLoader(buffer)
			result, err := schema.Validate(payload)
			if err != nil {
				badRequest(w, r, "invalid payload", err.Error())
				return
			}

			if !result.Valid() {
				badRequest(w, r, "invalid payload", "")
				return
			}

			// we need to reinject the body because it has been consumed previously
			r.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))
			next.ServeHTTP(w, r)
		})
	}
}
