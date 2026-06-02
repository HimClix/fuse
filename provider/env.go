package provider

import (
	"os"
	"reflect"
	"strings"

	"github.com/himclix/fuse"
)

type envProvider struct {
	prefix     string
	targetType reflect.Type
}

func (p *envProvider) Name() string { return "env" }

func (p *envProvider) Load() (map[string]any, error) {
	fields := fuse.GetStructMeta(p.targetType)
	result := make(map[string]any)

	for _, fm := range fields {
		envName := p.envName(fm)
		val, ok := os.LookupEnv(envName)
		if !ok {
			continue
		}
		setNestedKey(result, fuse.ConfigPath(fm.Path), val)
	}
	return result, nil
}

// envName returns the env var name for a field.
// Explicit conf:"env:NAME" takes priority; otherwise auto-derive from prefix + path.
func (p *envProvider) envName(fm fuse.FieldMeta) string {
	if fm.Conf.Env != "" {
		return fm.Conf.Env
	}
	name := strings.ToUpper(strings.ReplaceAll(fm.Path, ".", "_"))
	if p.prefix != "" {
		return p.prefix + "_" + name
	}
	return name
}

// Env returns a Provider that reads environment variables.
// Prefix is prepended to auto-derived var names (e.g., prefix "APP" + field "DB.Host" → APP_DB_HOST).
// Fields with explicit conf:"env:NAME" use that name directly (no prefix).
func Env[T any](prefix string) fuse.Provider {
	var zero T
	return &envProvider{
		prefix:     strings.ToUpper(prefix),
		targetType: reflect.TypeOf(zero),
	}
}
