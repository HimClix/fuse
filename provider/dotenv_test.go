package provider

import "testing"

func TestDotEnvProvider(t *testing.T) {
	content := "# comment\nDB_HOST=localhost\nDB_PORT=5432\nDB_PASSWORD=\"quoted\"\nSINGLE='single'\nEMPTY=\n"
	path := writeTempFile(t, ".env", content)

	m, err := DotEnv(path).Load()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		key, want string
	}{
		{"DB_HOST", "localhost"},
		{"DB_PORT", "5432"},
		{"DB_PASSWORD", "quoted"},
		{"SINGLE", "single"},
		{"EMPTY", ""},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := m[tt.key]
			if !ok {
				t.Fatalf("key %q not found", tt.key)
			}
			if got != tt.want {
				t.Errorf("%s = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestDotEnvSkipsComments(t *testing.T) {
	path := writeTempFile(t, ".env", "# comment\nKEY=value\n  # indented\n")
	m, err := DotEnv(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1 {
		t.Errorf("expected 1 key, got %d: %v", len(m), m)
	}
}

func TestDotEnvMissingFile(t *testing.T) {
	if _, err := DotEnv("/nonexistent/.env").Load(); err == nil {
		t.Error("expected error for missing file")
	}
}
