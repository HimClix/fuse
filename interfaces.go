package fuse

import (
	"reflect"

	"github.com/himclix/fuse/internal/tag"
)

// Provider loads configuration from a source into a key-value map.
type Provider interface {
	Name() string
	Load() (map[string]any, error)
}

// Parser converts raw file bytes into a nested key-value map.
type Parser interface {
	Parse(data []byte) (map[string]any, error)
}

// Result holds the loaded config and source metadata.
// Immutable after Load returns.
type Result[T any] struct {
	Config  *T
	Sources map[string]string // "db.host" → "config/prod.toml"
}

// TagInfo holds parsed conf:"..." directives. Re-exported from internal/tag.
type TagInfo = tag.Info

// FieldMeta holds metadata for a struct field. Re-exported from internal/tag.
type FieldMeta = tag.FieldMeta

// GetStructMeta returns cached field metadata for a struct type.
func GetStructMeta(t reflect.Type) []FieldMeta {
	return tag.Fields(t)
}

// ConfigPath returns the lowercase dot-separated config key for a field path.
func ConfigPath(path string) string {
	return tag.ConfigPath(path)
}
