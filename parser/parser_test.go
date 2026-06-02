package parser

import "testing"

func TestParsers(t *testing.T) {
	tests := []struct {
		name    string
		parser  func() interface{ Parse([]byte) (map[string]any, error) }
		input   string
		wantKey string
		wantVal any
	}{
		{"TOML", func() interface{ Parse([]byte) (map[string]any, error) } { return TOML() }, "[app]\nport = 8080", "app", map[string]any{"port": int64(8080)}},
		{"YAML", func() interface{ Parse([]byte) (map[string]any, error) } { return YAML() }, "app:\n  port: 8080", "app", map[string]any{"port": 8080}},
		{"JSON", func() interface{ Parse([]byte) (map[string]any, error) } { return JSON() }, `{"app":{"port":8080}}`, "app", map[string]any{"port": float64(8080)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.parser().Parse([]byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			if m[tt.wantKey] == nil {
				t.Errorf("key %q not found", tt.wantKey)
			}
		})
	}
}

func TestParsersInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		parser interface{ Parse([]byte) (map[string]any, error) }
		input  string
	}{
		{"TOML invalid", TOML(), "[broken"},
		{"YAML invalid", YAML(), ":\n  - :\n  bad"},
		{"JSON invalid", JSON(), "{broken"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.parser.Parse([]byte(tt.input)); err == nil {
				t.Error("expected error")
			}
		})
	}
}
