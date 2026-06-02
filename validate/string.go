package validate

import (
	"strings"
	"unicode"
)

func registerStringValidators(e *Engine) {
	e.validators["alpha"] = validateAlpha
	e.validators["alphanum"] = validateAlphanum
	e.validators["numeric"] = validateNumeric
	e.validators["lowercase"] = validateLowercase
	e.validators["uppercase"] = validateUppercase
	e.validators["contains"] = validateContains
	e.validators["excludes"] = validateExcludes
	e.validators["startswith"] = validateStartsWith
	e.validators["endswith"] = validateEndsWith
}

func validateAlpha(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func validateAlphanum(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func validateNumeric(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
	}
	if start >= len(s) {
		return false
	}
	hasDot := false
	for i := start; i < len(s); i++ {
		if s[i] == '.' && !hasDot {
			hasDot = true
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func validateLowercase(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return s == strings.ToLower(s) && s != ""
}

func validateUppercase(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return s == strings.ToUpper(s) && s != ""
}

func validateContains(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return strings.Contains(s, info.Param)
}

func validateExcludes(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return true
	}
	return !strings.Contains(s, info.Param)
}

func validateStartsWith(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return strings.HasPrefix(s, info.Param)
}

func validateEndsWith(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return strings.HasSuffix(s, info.Param)
}
