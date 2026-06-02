package provider

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/himclix/fuse"
)

type dotenvProvider struct {
	path string
}

func (p *dotenvProvider) Name() string { return p.path }

func (p *dotenvProvider) Load() (map[string]any, error) {
	f, err := os.Open(p.path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", p.path, err)
	}
	defer f.Close()

	result := make(map[string]any)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = unquote(val)

		result[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", p.path, err)
	}
	return result, nil
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// DotEnv returns a Provider that reads a .env file (KEY=VALUE format).
func DotEnv(path string) fuse.Provider {
	return &dotenvProvider{path: path}
}
