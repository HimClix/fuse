package validate

import (
	"reflect"
	"strings"
)

func registerCrossfieldValidators(e *Engine) {
	e.validators["required_if"] = validateRequiredIf
	e.validators["required_unless"] = validateRequiredUnless
	e.validators["required_with"] = validateRequiredWith
	e.validators["required_without"] = validateRequiredWithout
	e.validators["eqfield"] = validateEqField
	e.validators["nefield"] = validateNeField
	e.validators["gtfield"] = validateGtField
	e.validators["ltfield"] = validateLtField
}

func validateRequiredIf(info FieldInfo) bool {
	parts := strings.Fields(info.Param)
	if len(parts) < 2 {
		return true
	}
	otherField := parts[0]
	expectedVal := parts[1]

	otherValue := getFieldValue(info.Top, otherField)
	if otherValue == nil {
		return true
	}
	if valueToString(otherValue) != expectedVal {
		return true
	}
	return !isZeroValue(reflect.ValueOf(info.Value))
}

func validateRequiredUnless(info FieldInfo) bool {
	parts := strings.Fields(info.Param)
	if len(parts) < 2 {
		return true
	}
	otherField := parts[0]
	expectedVal := parts[1]

	otherValue := getFieldValue(info.Top, otherField)
	if otherValue == nil {
		return !isZeroValue(reflect.ValueOf(info.Value))
	}
	if valueToString(otherValue) == expectedVal {
		return true
	}
	return !isZeroValue(reflect.ValueOf(info.Value))
}

func validateRequiredWith(info FieldInfo) bool {
	otherField := strings.TrimSpace(info.Param)
	otherValue := getFieldValue(info.Top, otherField)
	if otherValue == nil || isZeroValue(reflect.ValueOf(otherValue)) {
		return true
	}
	return !isZeroValue(reflect.ValueOf(info.Value))
}

func validateRequiredWithout(info FieldInfo) bool {
	otherField := strings.TrimSpace(info.Param)
	otherValue := getFieldValue(info.Top, otherField)
	if otherValue != nil && !isZeroValue(reflect.ValueOf(otherValue)) {
		return true
	}
	return !isZeroValue(reflect.ValueOf(info.Value))
}

func validateEqField(info FieldInfo) bool {
	otherValue := getFieldValue(info.Top, info.Param)
	if otherValue == nil {
		return false
	}
	return reflect.DeepEqual(info.Value, otherValue)
}

func validateNeField(info FieldInfo) bool {
	otherValue := getFieldValue(info.Top, info.Param)
	if otherValue == nil {
		return true
	}
	return !reflect.DeepEqual(info.Value, otherValue)
}

func validateGtField(info FieldInfo) bool {
	return compareFields(info, func(a, b float64) bool { return a > b })
}

func validateLtField(info FieldInfo) bool {
	return compareFields(info, func(a, b float64) bool { return a < b })
}

func compareFields(info FieldInfo, cmp func(float64, float64) bool) bool {
	otherValue := getFieldValue(info.Top, info.Param)
	if otherValue == nil {
		return false
	}
	a := toFloat(info.Value)
	b := toFloat(otherValue)
	if a == nil || b == nil {
		return false
	}
	return cmp(*a, *b)
}

func getFieldValue(top any, fieldName string) any {
	if top == nil {
		return nil
	}
	rv := reflect.ValueOf(top)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	parts := strings.Split(fieldName, ".")
	for _, part := range parts {
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return nil
		}
		fv := rv.FieldByName(part)
		if !fv.IsValid() {
			return nil
		}
		rv = fv
	}
	return rv.Interface()
}

func toFloat(v any) *float64 {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f := float64(rv.Int())
		return &f
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f := float64(rv.Uint())
		return &f
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		return &f
	}
	return nil
}
