package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/himclix/fuse"
)

type fileProvider struct {
	path     string
	parser   fuse.Parser
	optional bool
}

func (p *fileProvider) Name() string { return p.path }

func (p *fileProvider) Load() (map[string]any, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if p.optional && os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", p.path, err)
	}
	m, err := p.parser.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", p.path, err)
	}
	return m, nil
}

// File returns a Provider that reads and parses a config file.
func File(path string, parser fuse.Parser) fuse.Provider {
	return &fileProvider{path: path, parser: parser}
}

// OptionalFile returns a Provider that reads a config file if it exists.
// Returns empty config (no error) if the file is missing.
func OptionalFile(path string, parser fuse.Parser) fuse.Provider {
	return &fileProvider{path: path, parser: parser, optional: true}
}

// AutoFile returns a Provider that auto-detects the parser from file extension.
// Requires parsers to be passed as a map: {"toml": parser.TOML(), "yaml": parser.YAML(), ...}
func AutoFile(path string, parsers map[string]fuse.Parser) fuse.Provider {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	p, ok := parsers[ext]
	if !ok {
		p, ok = parsers[strings.ToLower(ext)]
	}
	if !ok {
		return &errProvider{name: path, err: fmt.Errorf("no parser registered for .%s", ext)}
	}
	return &fileProvider{path: path, parser: p}
}

type errProvider struct {
	name string
	err  error
}

func (p *errProvider) Name() string                    { return p.name }
func (p *errProvider) Load() (map[string]any, error) { return nil, p.err }
