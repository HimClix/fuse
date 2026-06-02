package fuse

import (
	"strings"
	"testing"

	"github.com/himclix/fuse/internal/summary"
)

func TestSummary(t *testing.T) {
	type cfg struct {
		Port int    `conf:"default:8080"`
		Host string `conf:"default:localhost"`
		Pass string `conf:"secret"`
	}

	tests := []struct {
		name     string
		cfg      *cfg
		sources  map[string]string
		contains []string
		excludes []string
	}{
		{
			name:     "masks secrets",
			cfg:      &cfg{Port: 8080, Host: "localhost", Pass: "my-secret"},
			sources:  map[string]string{"port": "defaults", "pass": "env"},
			contains: []string{"*****", "defaults", "env"},
			excludes: []string{"my-secret"},
		},
		{
			name:     "shows sources",
			cfg:      &cfg{Port: 9090, Host: "api.com", Pass: "x"},
			sources:  map[string]string{"port": "config/prod.toml"},
			contains: []string{"config/prod.toml", "9090"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := summary.Format(tt.cfg, tt.sources)
			for _, want := range tt.contains {
				if !strings.Contains(s, want) {
					t.Errorf("summary missing %q:\n%s", want, s)
				}
			}
			for _, bad := range tt.excludes {
				if strings.Contains(s, bad) {
					t.Errorf("summary should NOT contain %q:\n%s", bad, s)
				}
			}
		})
	}
}

func TestSummaryEmpty(t *testing.T) {
	type empty struct{}
	s := summary.Format(&empty{}, map[string]string{})
	if !strings.Contains(s, "no config fields") {
		t.Error("expected 'no config fields'")
	}
}
