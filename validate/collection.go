package validate

import "reflect"

func registerCollectionValidators(e *Engine) {
	e.validators["unique"] = validateUnique
}

func validateUnique(info FieldInfo) bool {
	rv := reflect.ValueOf(info.Value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return true
	}

	seen := make(map[any]bool, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		val := rv.Index(i).Interface()
		if seen[val] {
			return false
		}
		seen[val] = true
	}
	return true
}
