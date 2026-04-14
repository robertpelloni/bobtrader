package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Validator defines a standard interface for input DTOs.
type Validator interface {
	Validate() error
}

var (
	// Basic regex for symbols (e.g. BTC, ETH-USDT)
	symbolRegex = regexp.MustCompile(`^[A-Z0-9-]{2,10}$`)
)

// DecodeAndValidate reads JSON from the body, unmarshals it into v, and calls Validate().
// It also enforces a maximum body size to prevent memory exhaustion attacks.
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, v Validator) error {
	// Max 1MB body size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Strict matching to prevent mass assignment

	err := dec.Decode(v)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("request body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("request body contains invalid value for field %q", unmarshalTypeError.Field)
		case strings.Contains(err.Error(), "unknown field"):
			return fmt.Errorf("request body contains unknown field")
		case err == io.EOF:
			return fmt.Errorf("request body must not be empty")
		case err.Error() == "http: request body too large":
			return fmt.Errorf("request body must not be larger than 1MB")
		default:
			return fmt.Errorf("failed to decode JSON: %v", err)
		}
	}

	// Ensure only one JSON object was in the body
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("request body must only contain a single JSON object")
	}

	if err := v.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
