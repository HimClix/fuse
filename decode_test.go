package fuse

import (
	"testing"
	"time"

	"github.com/himclix/fuse/internal/decode"
)

func TestDecodeScalars(t *testing.T) {
	type cfg struct {
		S string
		I int
		F float64
		B bool
		U uint
	}
	tests := []struct {
		name string
		data map[string]any
		check func(cfg) string
	}{
		{"string", map[string]any{"s": "hello"}, func(c cfg) string { if c.S != "hello" { return "S" }; return "" }},
		{"int", map[string]any{"i": int64(42)}, func(c cfg) string { if c.I != 42 { return "I" }; return "" }},
		{"int from string", map[string]any{"i": "99"}, func(c cfg) string { if c.I != 99 { return "I" }; return "" }},
		{"float", map[string]any{"f": 3.14}, func(c cfg) string { if c.F != 3.14 { return "F" }; return "" }},
		{"bool", map[string]any{"b": true}, func(c cfg) string { if !c.B { return "B" }; return "" }},
		{"bool from string", map[string]any{"b": "true"}, func(c cfg) string { if !c.B { return "B" }; return "" }},
		{"uint", map[string]any{"u": int64(10)}, func(c cfg) string { if c.U != 10 { return "U" }; return "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c cfg
			if err := decode.Map(tt.data, &c); err != nil {
				t.Fatal(err)
			}
			if field := tt.check(c); field != "" {
				t.Errorf("field %s has wrong value: %+v", field, c)
			}
		})
	}
}

func TestDecodeNested(t *testing.T) {
	type cfg struct {
		App struct{ Port int; Host string }
		DB  struct{ Host string }
	}
	data := map[string]any{
		"app": map[string]any{"port": int64(9090), "host": "api.test.com"},
		"db":  map[string]any{"host": "db.test.com"},
	}
	var c cfg
	if err := decode.Map(data, &c); err != nil {
		t.Fatal(err)
	}
	tests := []struct{ name, got, want string }{
		{"App.Host", c.App.Host, "api.test.com"},
		{"DB.Host", c.DB.Host, "db.test.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestDecodeSlice(t *testing.T) {
	var c struct{ Tags []string; Ports []int }
	data := map[string]any{"tags": []any{"go", "config"}, "ports": []any{int64(80), int64(443)}}
	if err := decode.Map(data, &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Tags) != 2 || c.Tags[0] != "go" {
		t.Errorf("Tags = %v", c.Tags)
	}
	if len(c.Ports) != 2 || c.Ports[1] != 443 {
		t.Errorf("Ports = %v", c.Ports)
	}
}

func TestDecodeMap(t *testing.T) {
	var c struct{ Labels map[string]string }
	data := map[string]any{"labels": map[string]any{"env": "prod", "team": "platform"}}
	if err := decode.Map(data, &c); err != nil {
		t.Fatal(err)
	}
	if c.Labels["env"] != "prod" || c.Labels["team"] != "platform" {
		t.Errorf("Labels = %v", c.Labels)
	}
}

func TestDecodeDuration(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  time.Duration
	}{
		{"string 5s", "5s", 5 * time.Second},
		{"string 100ms", "100ms", 100 * time.Millisecond},
		{"int seconds", int64(10), 10 * time.Second},
		{"float seconds", float64(2), 2 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c struct{ Timeout time.Duration }
			if err := decode.Map(map[string]any{"timeout": tt.input}, &c); err != nil {
				t.Fatal(err)
			}
			if c.Timeout != tt.want {
				t.Errorf("got %v, want %v", c.Timeout, tt.want)
			}
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		target any
	}{
		{"nil pointer", map[string]any{}, (*struct{ X int })(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := decode.Map(tt.data, tt.target); err == nil {
				t.Error("expected error")
			}
		})
	}
}
