package provider

import (
	"testing"

	"github.com/himclix/fuse"
)

type flagsTestConfig struct {
	Port int    `conf:"default:8080"`
	Host string `conf:"default:localhost"`
	Name string `conf:"flag:service-name"`
	DB   struct {
		Host string
		Port int
	}
}

func TestFlagsProvider(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantVals []string
	}{
		{"key=value", []string{"--port=9090"}, []string{"9090"}},
		{"key space value", []string{"--port", "9090"}, []string{"9090"}},
		{"explicit flag name", []string{"--service-name=my-app"}, []string{"my-app"}},
		{"nested key", []string{"--db-host=flag-db.com"}, []string{"flag-db.com"}},
		{"boolean flag", []string{"--unknown-flag"}, nil},
		{"empty args", nil, nil},
		{"multiple flags", []string{"--port=80", "--host=api.com"}, []string{"80", "api.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := Flags[flagsTestConfig](tt.args).Load()
			if err != nil {
				t.Fatal(err)
			}
			flat := flattenResult(m)
			for _, want := range tt.wantVals {
				found := false
				for _, got := range flat {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("want %q in result, got %v (raw: %v)", want, flat, m)
				}
			}
		})
	}
}

func TestFlagsProviderName(t *testing.T) {
	if Flags[flagsTestConfig](nil).Name() != "flags" {
		t.Error("Name() should be 'flags'")
	}
}

var _ fuse.Provider = Flags[flagsTestConfig](nil)
