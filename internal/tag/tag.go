package tag

import (
	"reflect"
	"strings"
	"sync"
)

// Info holds parsed conf:"..." directives for a single field.
type Info struct {
	Default    string
	HasDefault bool
	Env        string
	Flag       string
	Secret     bool
}

// FieldMeta holds all metadata for a single struct field.
type FieldMeta struct {
	Name     string
	Path     string // dot-separated: "DB.Host"
	Index    []int  // reflect field index chain
	Type     reflect.Type
	Kind     reflect.Kind
	Conf     Info
	Validate string // raw validate:"..." tag
}

// StructMeta holds cached metadata for an entire struct type.
type StructMeta struct {
	Fields []FieldMeta
	ByPath map[string]*FieldMeta
}

var cache sync.Map

// Get returns cached metadata for a struct type.
func Get(t reflect.Type) *StructMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if cached, ok := cache.Load(t); ok {
		return cached.(*StructMeta)
	}
	meta := build(t, "", nil)
	cache.Store(t, meta)
	return meta
}

// Fields returns the field list for a struct type (convenience for providers).
func Fields(t reflect.Type) []FieldMeta {
	return Get(t).Fields
}

func build(t reflect.Type, prefix string, indexChain []int) *StructMeta {
	meta := &StructMeta{ByPath: make(map[string]*FieldMeta)}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		idx := append(append([]int{}, indexChain...), i)
		ft := f.Type
		kind := ft.Kind()

		if kind == reflect.Ptr {
			ft = ft.Elem()
			kind = ft.Kind()
		}

		path := f.Name
		if prefix != "" {
			path = prefix + "." + f.Name
		}

		if kind == reflect.Struct && !IsSpecialType(ft) {
			_, hasConf := f.Tag.Lookup("conf")
			_, hasValidate := f.Tag.Lookup("validate")
			if !hasConf && !hasValidate {
				nested := build(ft, path, idx)
				meta.Fields = append(meta.Fields, nested.Fields...)
				for k, v := range nested.ByPath {
					meta.ByPath[k] = v
				}
				continue
			}
		}

		fm := FieldMeta{
			Name:     f.Name,
			Path:     path,
			Index:    idx,
			Type:     ft,
			Kind:     kind,
			Conf:     ParseConf(f.Tag.Get("conf")),
			Validate: f.Tag.Get("validate"),
		}

		meta.Fields = append(meta.Fields, fm)
		meta.ByPath[path] = &meta.Fields[len(meta.Fields)-1]
	}
	return meta
}

// ParseConf parses a conf:"..." struct tag into Info.
func ParseConf(raw string) Info {
	var ti Info
	if raw == "" || raw == "-" {
		return ti
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == "secret" {
			ti.Secret = true
			continue
		}
		key, val, hasVal := strings.Cut(part, ":")
		if !hasVal {
			continue
		}
		switch key {
		case "default":
			ti.Default = val
			ti.HasDefault = true
		case "env":
			ti.Env = val
		case "flag":
			ti.Flag = val
		}
	}
	return ti
}

// ConfigPath returns the lowercase dot-separated config key.
func ConfigPath(path string) string {
	parts := strings.Split(path, ".")
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	return strings.Join(parts, ".")
}

// IsSpecialType returns true for time.Time and time.Duration.
func IsSpecialType(t reflect.Type) bool {
	return t.PkgPath() == "time" && (t.Name() == "Time" || t.Name() == "Duration")
}
