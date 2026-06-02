package validate

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidatorFunc validates a single field. Return true if valid.
type ValidatorFunc func(info FieldInfo) bool

// FieldInfo provides context to validators about the field being validated.
type FieldInfo struct {
	Value  any          // current field value
	Field  string       // struct field name: "Port"
	Path   string       // full dotted path: "Server.Port"
	Param  string       // tag parameter: "8" from "min=8"
	Kind   reflect.Kind // reflect.Int, reflect.String, etc.
	Type   reflect.Type // full reflect type
	Parent any          // parent struct (for cross-field)
	Top    any          // top-level struct (for cross-field)
}

// FieldError represents a single field's validation failure.
type FieldError struct {
	Field   string // struct field name
	Path    string // full dotted path
	Tag     string // validation tag that failed
	Param   string // tag parameter
	Value   any    // current field value
	Message string // human-readable message
}

func (fe FieldError) String() string {
	if fe.Message != "" {
		return fe.Message
	}
	if fe.Param != "" {
		return fmt.Sprintf("field %q failed on %q (param=%q, value=%v)", fe.Path, fe.Tag, fe.Param, fe.Value)
	}
	return fmt.Sprintf("field %q failed on %q (value=%v)", fe.Path, fe.Tag, fe.Value)
}

func (fe FieldError) Error() string { return fe.String() }

// ValidationErrors collects all validation failures.
type ValidationErrors []FieldError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation passed"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d validation error(s):\n", len(ve))
	for _, fe := range ve {
		fmt.Fprintf(&b, "  - %s\n", fe.String())
	}
	return b.String()
}

// validationTag holds a parsed single validation directive.
type validationTag struct {
	Name  string // "required", "min", "email", etc.
	Param string // "8" from "min=8", "a b c" from "oneof=a b c"
}

// parseValidateTag parses "required,min=8,oneof=a b c" into individual tags.
// Supports AND (comma) and OR (pipe) combinators.
func parseValidateTag(raw string) [][]validationTag {
	if raw == "" || raw == "-" {
		return nil
	}

	// Split by comma for AND groups
	// Each AND group may contain pipe-separated OR alternatives
	andParts := strings.Split(raw, ",")
	var groups [][]validationTag

	for _, andPart := range andParts {
		andPart = strings.TrimSpace(andPart)
		if andPart == "" {
			continue
		}

		// Split by pipe for OR within this AND group
		orParts := strings.Split(andPart, "|")
		var orGroup []validationTag
		for _, orPart := range orParts {
			orPart = strings.TrimSpace(orPart)
			if orPart == "" {
				continue
			}
			name, param, _ := strings.Cut(orPart, "=")
			orGroup = append(orGroup, validationTag{Name: name, Param: param})
		}
		if len(orGroup) > 0 {
			groups = append(groups, orGroup)
		}
	}
	return groups
}
