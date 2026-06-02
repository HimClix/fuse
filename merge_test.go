package fuse

import (
	"testing"

	"github.com/himclix/fuse/internal/merge"
)

func TestMergeAll(t *testing.T) {
	tests := []struct {
		name       string
		results    []merge.Result
		wantKey    string
		wantVal    any
		wantSource string
	}{
		{
			name: "later scalar overwrites earlier",
			results: []merge.Result{
				{Name: "base", Data: map[string]any{"port": 8080}},
				{Name: "override", Data: map[string]any{"port": 9090}},
			},
			wantKey: "port", wantVal: 9090, wantSource: "override",
		},
		{
			name: "earlier value preserved if not overridden",
			results: []merge.Result{
				{Name: "base", Data: map[string]any{"host": "localhost"}},
				{Name: "override", Data: map[string]any{"port": 9090}},
			},
			wantKey: "host", wantVal: "localhost", wantSource: "base",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, sources := merge.All(tt.results)
			if merged[tt.wantKey] != tt.wantVal {
				t.Errorf("%s = %v, want %v", tt.wantKey, merged[tt.wantKey], tt.wantVal)
			}
			if sources[tt.wantKey] != tt.wantSource {
				t.Errorf("source[%s] = %q, want %q", tt.wantKey, sources[tt.wantKey], tt.wantSource)
			}
		})
	}
}

func TestMergeNestedMaps(t *testing.T) {
	results := []merge.Result{
		{Name: "base", Data: map[string]any{"db": map[string]any{"host": "localhost", "port": 5432}}},
		{Name: "prod", Data: map[string]any{"db": map[string]any{"host": "prod-db"}}},
	}
	merged, sources := merge.All(results)

	db := merged["db"].(map[string]any)
	tests := []struct {
		key, wantSource string
		wantVal         any
	}{
		{"host", "prod", "prod-db"},
		{"port", "base", 5432},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if db[tt.key] != tt.wantVal {
				t.Errorf("db.%s = %v, want %v", tt.key, db[tt.key], tt.wantVal)
			}
			if sources["db."+tt.key] != tt.wantSource {
				t.Errorf("source[db.%s] = %q, want %q", tt.key, sources["db."+tt.key], tt.wantSource)
			}
		})
	}
}

func TestLookupNested(t *testing.T) {
	m := map[string]any{"db": map[string]any{"Host": "localhost"}}

	tests := []struct {
		path    string
		wantVal any
		wantOk  bool
	}{
		{"db.Host", "localhost", true},
		{"db.host", "localhost", true}, // case-insensitive
		{"db.missing", nil, false},
		{"nonexistent.key", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			v, ok := merge.LookupNested(m, tt.path)
			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && v != tt.wantVal {
				t.Errorf("value = %v, want %v", v, tt.wantVal)
			}
		})
	}
}
