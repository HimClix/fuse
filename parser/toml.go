package parser

import (
	"github.com/himclix/fuse"
	"github.com/pelletier/go-toml/v2"
)

type tomlParser struct{}

func (p *tomlParser) Parse(data []byte) (map[string]any, error) {
	var m map[string]any
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// TOML returns a fuse.Parser for TOML files.
func TOML() fuse.Parser { return &tomlParser{} }
