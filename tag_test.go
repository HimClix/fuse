package fuse

import (
	"reflect"
	"testing"

	"github.com/himclix/fuse/internal/tag"
)

func TestParseConfTag(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want tag.Info
	}{
		{"full tag", "default:8080,env:APP_PORT,flag:port", tag.Info{Default: "8080", HasDefault: true, Env: "APP_PORT", Flag: "port"}},
		{"env and secret", "env:DB_PASS,secret", tag.Info{Env: "DB_PASS", Secret: true}},
		{"default only", "default:myapp", tag.Info{Default: "myapp", HasDefault: true}},
		{"skip tag", "-", tag.Info{}},
		{"empty tag", "", tag.Info{}},
		{"secret only", "secret", tag.Info{Secret: true}},
		{"whitespace", " default:x , env:Y ", tag.Info{Default: "x", HasDefault: true, Env: "Y"}},
		{"flag only", "flag:db-host", tag.Info{Flag: "db-host"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tag.ParseConf(tt.raw)
			if got != tt.want {
				t.Errorf("ParseConf(%q)\n  got  %+v\n  want %+v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	tests := []struct {
		path, want string
	}{
		{"DB.Host", "db.host"},
		{"App.Server.Port", "app.server.port"},
		{"Name", "name"},
		{"A.B.C.D", "a.b.c.d"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := tag.ConfigPath(tt.path); got != tt.want {
				t.Errorf("ConfigPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestStructMetaNestedFields(t *testing.T) {
	type nested struct {
		App struct {
			Port int    `conf:"default:8080"`
			Host string `conf:"default:localhost"`
		}
		DB struct {
			Host string `conf:"env:DB_HOST"`
		}
	}

	meta := tag.Get(reflect.TypeOf(nested{}))
	want := []string{"App.Port", "App.Host", "DB.Host"}
	paths := make(map[string]bool)
	for _, f := range meta.Fields {
		paths[f.Path] = true
	}
	for _, p := range want {
		if !paths[p] {
			t.Errorf("missing field path %q", p)
		}
	}
}

func TestStructMetaCache(t *testing.T) {
	type cfg struct{ Port int }
	typ := reflect.TypeOf(cfg{})
	if tag.Get(typ) != tag.Get(typ) {
		t.Error("cache should return same pointer")
	}
}
