package provider

import (
	"reflect"
	"strings"

	"github.com/himclix/fuse"
)

type defaultsProvider struct {
	targetType reflect.Type
}

func (p *defaultsProvider) Name() string { return "defaults" }

func (p *defaultsProvider) Load() (map[string]any, error) {
	fields := fuse.GetStructMeta(p.targetType)
	result := make(map[string]any)

	for _, fm := range fields {
		if fm.Conf.HasDefault {
			setNestedKey(result, fuse.ConfigPath(fm.Path), fm.Conf.Default)
		}
	}
	return result, nil
}

// Defaults returns a Provider that populates values from conf:"default:..." tags.
func Defaults[T any]() fuse.Provider {
	var zero T
	return &defaultsProvider{targetType: reflect.TypeOf(zero)}
}

// setNestedKey sets a value in a nested map by dot-separated path.
func setNestedKey(m map[string]any, path string, val any) {
	parts := strings.Split(path, ".")
	current := m
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = val
			return
		}
		sub, ok := current[part].(map[string]any)
		if !ok {
			sub = make(map[string]any)
			current[part] = sub
		}
		current = sub
	}
}
