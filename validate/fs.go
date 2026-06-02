package validate

import (
	"os"
	"path/filepath"
)

func registerFSValidators(e *Engine) {
	e.validators["dir"] = validateDir
	e.validators["dirpath"] = validateDirPath
	e.validators["file"] = validateFile
	e.validators["filepath"] = validateFilePath
}

func validateDir(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	fi, err := os.Stat(s)
	return err == nil && fi.IsDir()
}

func validateDirPath(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	return filepath.IsAbs(s) || s == "." || s == ".." ||
		(len(s) > 0 && s[0] != 0)
}

func validateFile(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	fi, err := os.Stat(s)
	return err == nil && !fi.IsDir()
}

func validateFilePath(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	_, err := filepath.Abs(s)
	return err == nil && len(s) > 0
}
