package fuse

import (
	"fmt"

	"github.com/himclix/fuse/validate"
)

// ValidationErrors is an alias for validate.ValidationErrors.
// Returned by Load when validation fails. Iterate to inspect individual errors.
type ValidationErrors = validate.ValidationErrors

// FieldError is an alias for validate.FieldError.
// Represents a single field's validation failure.
type FieldError = validate.FieldError

// ConfigError represents a failure during config loading (provider/parse/decode).
type ConfigError struct {
	Provider string
	Err      error
}

func (e *ConfigError) Error() string {
	if e.Provider != "" {
		return fmt.Sprintf("fuse: provider %q: %v", e.Provider, e.Err)
	}
	return fmt.Sprintf("fuse: %v", e.Err)
}

func (e *ConfigError) Unwrap() error { return e.Err }
