package validate

import (
	"reflect"
	"strconv"
	"strings"
)

func registerCompareValidators(e *Engine) {
	e.validators["required"] = validateRequired
	e.validators["eq"] = validateEq
	e.validators["ne"] = validateNe
	e.validators["gt"] = validateGt
	e.validators["gte"] = validateGte
	e.validators["lt"] = validateLt
	e.validators["lte"] = validateLte
	e.validators["min"] = validateMin
	e.validators["max"] = validateMax
	e.validators["len"] = validateLen
	e.validators["oneof"] = validateOneof
}

func validateRequired(info FieldInfo) bool {
	v := reflect.ValueOf(info.Value)
	if !v.IsValid() {
		return false
	}
	return !isZeroValue(v)
}

func validateEq(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a == b })
}

func validateNe(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a != b })
}

func validateGt(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a > b })
}

func validateGte(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a >= b })
}

func validateLt(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a < b })
}

func validateLte(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a <= b })
}

func validateMin(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a >= b })
}

func validateMax(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a <= b })
}

func validateLen(info FieldInfo) bool {
	return compareNumOrLen(info, func(a, b float64) bool { return a == b })
}

func validateOneof(info FieldInfo) bool {
	allowed := strings.Fields(info.Param)
	val := valueToString(info.Value)
	for _, a := range allowed {
		if val == a {
			return true
		}
	}
	return false
}

// compareNumOrLen compares field value (numeric) or length (string/slice/map) against param.
func compareNumOrLen(info FieldInfo, cmp func(float64, float64) bool) bool {
	param, err := strconv.ParseFloat(info.Param, 64)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(info.Value)
	switch info.Kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp(float64(v.Int()), param)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cmp(float64(v.Uint()), param)
	case reflect.Float32, reflect.Float64:
		return cmp(v.Float(), param)
	case reflect.String:
		return cmp(float64(v.Len()), param)
	case reflect.Slice, reflect.Array, reflect.Map:
		return cmp(float64(v.Len()), param)
	}
	return false
}

func valueToString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return strings.TrimSpace(strings.Replace(
			reflect.ValueOf(v).String(), "<", "", -1))
	}
}
