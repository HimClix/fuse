package resolve

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/himclix/fuse/internal/merge"
	"github.com/himclix/fuse/internal/tag"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// Provider is the internal provider interface used by Boot.
type Provider interface {
	Name() string
	Load() (map[string]any, error)
}

// Defaults reads default values from conf:"default:..." tags.
func Defaults(targetType reflect.Type) Provider {
	return &defaultsProvider{targetType: targetType}
}

type defaultsProvider struct{ targetType reflect.Type }

func (p *defaultsProvider) Name() string { return "defaults" }
func (p *defaultsProvider) Load() (map[string]any, error) {
	fields := tag.Fields(p.targetType)
	result := make(map[string]any)
	for _, fm := range fields {
		if fm.Conf.HasDefault {
			merge.SetNested(result, tag.ConfigPath(fm.Path), fm.Conf.Default)
		}
	}
	return result, nil
}

// EnvVars reads environment variables with prefix + explicit env tags.
func EnvVars(targetType reflect.Type, prefix string) Provider {
	return &envProvider{prefix: strings.ToUpper(prefix), targetType: targetType}
}

type envProvider struct {
	prefix     string
	targetType reflect.Type
}

func (p *envProvider) Name() string { return "env" }
func (p *envProvider) Load() (map[string]any, error) {
	fields := tag.Fields(p.targetType)
	result := make(map[string]any)
	for _, fm := range fields {
		name := p.envName(fm)
		val, ok := os.LookupEnv(name)
		if !ok {
			continue
		}
		merge.SetNested(result, tag.ConfigPath(fm.Path), val)
	}
	return result, nil
}

func (p *envProvider) envName(fm tag.FieldMeta) string {
	if fm.Conf.Env != "" {
		return fm.Conf.Env
	}
	name := strings.ToUpper(strings.ReplaceAll(fm.Path, ".", "_"))
	if p.prefix != "" {
		return p.prefix + "_" + name
	}
	return name
}

// File reads and parses a config file.
func File(path, format string, optional bool) Provider {
	return &fileProvider{path: path, format: format, optional: optional}
}

type fileProvider struct {
	path     string
	format   string
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
	return ParseBytes(data, p.format)
}

// ParseBytes parses raw bytes in the given format.
func ParseBytes(data []byte, format string) (map[string]any, error) {
	var m map[string]any
	var err error
	switch strings.ToLower(format) {
	case "toml":
		err = toml.Unmarshal(data, &m)
	case "yaml", "yml":
		err = yaml.Unmarshal(data, &m)
	case "json":
		err = json.Unmarshal(data, &m)
	default:
		return nil, fmt.Errorf("unsupported format: %q", format)
	}
	return m, err
}
