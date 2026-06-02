package decode

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Map fills a struct from a config map.
func Map(data map[string]any, target any) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct, got %s", rv.Kind())
	}
	return decodeStruct(data, rv, "")
}

func decodeStruct(data map[string]any, rv reflect.Value, prefix string) error {
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
				fv.Set(reflect.New(ft.Elem()))
			}
			fv = fv.Elem()
			ft = ft.Elem()
			kind = ft.Kind()
		}

		fieldName := strings.ToLower(field.Name)
		val, found := caseInsensitiveLookup(data, fieldName)

		if kind == reflect.Struct && !isSpecialType(ft) && !found {
			subMap := findSubMap(data, fieldName)
			if subMap != nil {
				if err := decodeStruct(subMap, fv, joinPath(prefix, field.Name)); err != nil {
					return err
				}
			}
			continue
		}

		if kind == reflect.Struct && !isSpecialType(ft) {
			if subMap, ok := val.(map[string]any); ok {
				if err := decodeStruct(subMap, fv, joinPath(prefix, field.Name)); err != nil {
					return err
				}
				continue
			}
		}

		if !found {
			continue
		}

		if err := setFieldValue(fv, val, joinPath(prefix, field.Name)); err != nil {
			return fmt.Errorf("field %s: %w", joinPath(prefix, field.Name), err)
		}
	}
	return nil
}

func setFieldValue(fv reflect.Value, val any, path string) error {
	if val == nil {
		return nil
	}

	ft := fv.Type()
	kind := ft.Kind()

	if ft == reflect.TypeOf(time.Duration(0)) {
		return setDuration(fv, val)
	}
	if ft == reflect.TypeOf(time.Time{}) {
		return setTime(fv, val)
	}

	switch kind {
	case reflect.String:
		s, err := toString(val)
		if err != nil {
			return err
		}
		fv.SetString(s)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := toInt64(val)
		if err != nil {
			return err
		}
		fv.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := toUint64(val)
		if err != nil {
			return err
		}
		fv.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := toFloat64(val)
		if err != nil {
			return err
		}
		fv.SetFloat(f)

	case reflect.Bool:
		b, err := toBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(b)

	case reflect.Slice:
		return setSlice(fv, val, path)

	case reflect.Map:
		return setMap(fv, val, path)

	default:
		rval := reflect.ValueOf(val)
		if rval.Type().ConvertibleTo(ft) {
			fv.Set(rval.Convert(ft))
		} else {
			return fmt.Errorf("cannot convert %T to %s", val, ft)
		}
	}
	return nil
}

func setSlice(fv reflect.Value, val any, path string) error {
	slice, ok := val.([]any)
	if !ok {
		if s, ok := val.(string); ok {
			parts := strings.Split(s, ",")
			slice = make([]any, len(parts))
			for i, p := range parts {
				slice[i] = strings.TrimSpace(p)
			}
		} else {
			return fmt.Errorf("expected slice, got %T", val)
		}
	}

	elemType := fv.Type().Elem()
	result := reflect.MakeSlice(fv.Type(), len(slice), len(slice))
	for i, item := range slice {
		elem := reflect.New(elemType).Elem()
		if err := setFieldValue(elem, item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
		result.Index(i).Set(elem)
	}
	fv.Set(result)
	return nil
}

func setMap(fv reflect.Value, val any, path string) error {
	srcMap, ok := val.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map, got %T", val)
	}
	ft := fv.Type()
	result := reflect.MakeMap(ft)
	valType := ft.Elem()
	for k, v := range srcMap {
		mapVal := reflect.New(valType).Elem()
		if err := setFieldValue(mapVal, v, path+"."+k); err != nil {
			return fmt.Errorf("key %q: %w", k, err)
		}
		result.SetMapIndex(reflect.ValueOf(k), mapVal)
	}
	fv.Set(result)
	return nil
}

func setDuration(fv reflect.Value, val any) error {
	switch v := val.(type) {
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			n, nerr := strconv.ParseInt(v, 10, 64)
			if nerr != nil {
				return fmt.Errorf("invalid duration %q: %w", v, err)
			}
			d = time.Duration(n) * time.Second
		}
		fv.Set(reflect.ValueOf(d))
	case int64:
		fv.Set(reflect.ValueOf(time.Duration(v) * time.Second))
	case float64:
		fv.Set(reflect.ValueOf(time.Duration(v) * time.Second))
	default:
		n, err := toInt64(val)
		if err != nil {
			return fmt.Errorf("cannot convert %T to Duration", val)
		}
		fv.Set(reflect.ValueOf(time.Duration(n) * time.Second))
	}
	return nil
}

func setTime(fv reflect.Value, val any) error {
	s, ok := val.(string)
	if !ok {
		return fmt.Errorf("time.Time requires string, got %T", val)
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02"} {
		t, err := time.Parse(layout, s)
		if err == nil {
			fv.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return fmt.Errorf("cannot parse time %q", s)
}

func toString(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case fmt.Stringer:
		return val.String(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func toInt64(v any) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	case int32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

func toUint64(v any) (uint64, error) {
	n, err := toInt64(v)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("negative value %d for unsigned field", n)
	}
	return uint64(n), nil
}

func toFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func toBool(v any) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	case int:
		return val != 0, nil
	case int64:
		return val != 0, nil
	case float64:
		return val != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

func caseInsensitiveLookup(m map[string]any, key string) (any, bool) {
	if v, ok := m[key]; ok {
		return v, true
	}
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

func findSubMap(data map[string]any, name string) map[string]any {
	v, found := caseInsensitiveLookup(data, name)
	if !found {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func isSpecialType(t reflect.Type) bool {
	return t.PkgPath() == "time" && (t.Name() == "Time" || t.Name() == "Duration")
}
