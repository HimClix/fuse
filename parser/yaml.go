package parser

import (
	"github.com/himclix/fuse"
	"gopkg.in/yaml.v3"
)

type yamlParser struct{}

func (p *yamlParser) Parse(data []byte) (map[string]any, error) {
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// YAML returns a fuse.Parser for YAML files.
func YAML() fuse.Parser { return &yamlParser{} }
