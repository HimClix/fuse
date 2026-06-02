package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/himclix/fuse/parser"
)

func TestFileProvider(t *testing.T) {
	tests := []struct {
		name    string
		content string
		parser  interface {
			Parse([]byte) (map[string]any, error)
		}
		wantKey string
	}{
		{"TOML", "[app]\nport = 9090", parser.TOML(), "app"},
		{"YAML", "app:\n  port: 9090", parser.YAML(), "app"},
		{"JSON", `{"app":{"port":9090}}`, parser.JSON(), "app"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, "config", tt.content)
			m, err := File(path, tt.parser).Load()
			if err != nil {
				t.Fatal(err)
			}
			if m[tt.wantKey] == nil {
				t.Errorf("key %q not found in %v", tt.wantKey, m)
			}
		})
	}
}

func TestFileProviderErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"missing file", "/nonexistent/config.toml", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := File(tt.path, parser.TOML()).Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptionalFile(t *testing.T) {
	tests := []struct {
		name    string
		exists  bool
		wantNil bool
	}{
		{"missing returns nil", false, true},
		{"existing returns data", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/nonexistent/x.toml"
			if tt.exists {
				path = writeTempFile(t, "config.toml", "port = 8080")
			}
			m, err := OptionalFile(path, parser.TOML()).Load()
			if err != nil {
				t.Fatal(err)
			}
			if (m == nil) != tt.wantNil {
				t.Errorf("result nil = %v, want %v", m == nil, tt.wantNil)
			}
		})
	}
}

func TestFileProviderName(t *testing.T) {
	if File("config/prod.toml", parser.TOML()).Name() != "config/prod.toml" {
		t.Error("Name() should return path")
	}
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
