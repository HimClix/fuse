package parser

import (
	"encoding/json"

	"github.com/himclix/fuse"
)

type jsonParser struct{}

func (p *jsonParser) Parse(data []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// JSON returns a fuse.Parser for JSON files.
func JSON() fuse.Parser { return &jsonParser{} }
