package provider

import (
	"testing"

	"github.com/himclix/fuse"
)

type envTestConfig struct {
	Port int    `conf:"default:8080"`
	Host string `conf:"env:DB_HOST"`
}

func TestEnvProvider(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		envVars  map[string]string
		wantKeys int
	}{
		{"with prefix", "TEST", map[string]string{"TEST_PORT": "9090"}, 1},
		{"explicit env tag", "APP", map[string]string{"DB_HOST": "explicit.com"}, 1},
		{"no env set", "TEST", map[string]string{}, 0},
		{"empty prefix", "", map[string]string{"PORT": "3000"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			m, err := Env[envTestConfig](tt.prefix).Load()
			if err != nil {
				t.Fatal(err)
			}
			flat := flattenResult(m)
			if len(flat) != tt.wantKeys {
				t.Errorf("got %d keys, want %d: %v", len(flat), tt.wantKeys, m)
			}
		})
	}
}

func TestEnvProviderName(t *testing.T) {
	if Env[envTestConfig]("APP").Name() != "env" {
		t.Error("Name() should be 'env'")
	}
}

var _ fuse.Provider = Env[envTestConfig]("")

func flattenResult(m map[string]any) []string {
	var vals []string
	for _, v := range m {
		switch val := v.(type) {
		case string:
			vals = append(vals, val)
		case map[string]any:
			vals = append(vals, flattenResult(val)...)
		}
	}
	return vals
}
