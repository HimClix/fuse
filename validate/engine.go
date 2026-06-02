package validate

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Engine holds registered validators and validates structs.
type Engine struct {
	mu         sync.RWMutex
	validators map[string]ValidatorFunc
}

// New creates an Engine with all built-in validators registered.
func New() *Engine {
	e := &Engine{validators: make(map[string]ValidatorFunc)}
	e.registerBuiltins()
	return e
}

// Register adds a custom validator. Overwrites if tag already exists.
func (e *Engine) Register(tag string, fn ValidatorFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.validators[tag] = fn
}

func (e *Engine) getValidator(tag string) (ValidatorFunc, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	fn, ok := e.validators[tag]
	return fn, ok
}

// Validate validates a struct against its validate:"..." tags.
// Returns nil if valid, or ValidationErrors with all failures.
func (e *Engine) Validate(v any) ValidationErrors {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ValidationErrors{{Path: "", Tag: "", Message: "validate: expected struct"}}
	}

	var errs ValidationErrors
	e.validateStruct(rv, "", v, v, &errs)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (e *Engine) validateStruct(rv reflect.Value, prefix string, parent, top any, errs *ValidationErrors) {
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		fv := rv.Field(i)
		ft := field.Type
		kind := ft.Kind()

		if kind == reflect.Ptr {
			if fv.IsNil() {
				// Check required on nil pointer
				vtag := field.Tag.Get("validate")
				if vtag != "" && containsRequired(vtag) {
					path := joinPath(prefix, field.Name)
					*errs = append(*errs, FieldError{
						Field: field.Name, Path: path, Tag: "required",
						Value: nil, Message: "field \"" + path + "\" is required but is nil",
					})
				}
				continue
			}
			fv = fv.Elem()
			ft = ft.Elem()
			kind = ft.Kind()
		}

		path := joinPath(prefix, field.Name)

		// Recurse into nested structs
		if kind == reflect.Struct && !isSpecialType(ft) {
			vtag := field.Tag.Get("validate")
			if vtag == "" {
				e.validateStruct(fv, path, rv.Addr().Interface(), top, errs)
				continue
			}
		}

		vtag := field.Tag.Get("validate")
		if vtag == "" || vtag == "-" {
			continue
		}

		groups := parseValidateTag(vtag)
		e.validateField(fv, field, path, groups, parent, top, errs)
	}
}

func (e *Engine) validateField(fv reflect.Value, sf reflect.StructField, path string, groups [][]validationTag, parent, top any, errs *ValidationErrors) {
	ft := fv.Type()
	kind := ft.Kind()
	val := fv.Interface()

	// Split groups at "dive" boundary: pre-dive runs on parent, post-dive on each element
	preGroups, postGroups := splitAtDive(groups)

	for _, orGroup := range preGroups {
		if len(orGroup) == 1 {
			tag := orGroup[0]

			if tag.Name == "omitempty" {
				if isZeroValue(fv) {
					return
				}
				continue
			}

			fn, ok := e.getValidator(tag.Name)
			if !ok {
				continue
			}

			info := FieldInfo{
				Value:  val,
				Field:  sf.Name,
				Path:   path,
				Param:  tag.Param,
				Kind:   kind,
				Type:   ft,
				Parent: parent,
				Top:    top,
			}

			if !fn(info) {
				*errs = append(*errs, FieldError{
					Field: sf.Name,
					Path:  path,
					Tag:   tag.Name,
					Param: tag.Param,
					Value: val,
				})
			}
		} else {
			anyPassed := false
			for _, tag := range orGroup {
				fn, ok := e.getValidator(tag.Name)
				if !ok {
					continue
				}
				info := FieldInfo{
					Value: val, Field: sf.Name, Path: path,
					Param: tag.Param, Kind: kind, Type: ft,
					Parent: parent, Top: top,
				}
				if fn(info) {
					anyPassed = true
					break
				}
			}
			if !anyPassed {
				tagNames := make([]string, len(orGroup))
				for i, t := range orGroup {
					tagNames[i] = t.Name
				}
				*errs = append(*errs, FieldError{
					Field:   sf.Name,
					Path:    path,
					Tag:     strings.Join(tagNames, "|"),
					Value:   val,
					Message: "field \"" + path + "\" failed on \"" + strings.Join(tagNames, "|") + "\"",
				})
			}
		}
	}

	// Dive: apply post-dive tags to each slice/map element
	if len(postGroups) > 0 && (kind == reflect.Slice || kind == reflect.Array) {
		for i := 0; i < fv.Len(); i++ {
			elem := fv.Index(i)
			elemPath := path + "[" + itoa(i) + "]"
			e.validateField(elem, sf, elemPath, postGroups, parent, top, errs)
		}
	}
}

func (e *Engine) registerBuiltins() {
	registerCompareValidators(e)
	registerStringValidators(e)
	registerFormatValidators(e)
	registerCrossfieldValidators(e)
	registerCollectionValidators(e)
	registerFSValidators(e)
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
	return false
}

func containsRequired(tag string) bool {
	for _, part := range strings.Split(tag, ",") {
		if strings.TrimSpace(part) == "required" {
			return true
		}
	}
	return false
}

// splitAtDive splits validation groups into pre-dive (applied to parent)
// and post-dive (applied to each element). If no dive, post is empty.
func splitAtDive(groups [][]validationTag) (pre, post [][]validationTag) {
	for i, g := range groups {
		if len(g) == 1 && g[0].Name == "dive" {
			return groups[:i], groups[i+1:]
		}
	}
	return groups, nil
}

func isSpecialType(t reflect.Type) bool {
	return t.PkgPath() == "time" && (t.Name() == "Time" || t.Name() == "Duration")
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
