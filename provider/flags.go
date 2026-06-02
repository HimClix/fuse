package provider

import (
	"reflect"
	"strings"

	"github.com/himclix/fuse"
)

type flagsProvider struct {
	args       []string
	targetType reflect.Type
}

func (p *flagsProvider) Name() string { return "flags" }

func (p *flagsProvider) Load() (map[string]any, error) {
	parsed := parseArgs(p.args)
	if len(parsed) == 0 {
		return nil, nil
	}

	fields := fuse.GetStructMeta(p.targetType)
	flagMap := buildFlagMap(fields)
	result := make(map[string]any)

	for key, val := range parsed {
		confPath, ok := flagMap[strings.ToLower(key)]
		if !ok {
			continue
		}
		setNestedKey(result, confPath, val)
	}
	return result, nil
}

// buildFlagMap creates a mapping from flag name → config path.
// Uses explicit conf:"flag:name" or auto-derives from field path (kebab-case).
func buildFlagMap(fields []fuse.FieldMeta) map[string]string {
	m := make(map[string]string)
	for _, fm := range fields {
		confPath := fuse.ConfigPath(fm.Path)
		if fm.Conf.Flag != "" {
			m[strings.ToLower(fm.Conf.Flag)] = confPath
		} else {
			flagName := strings.ToLower(strings.ReplaceAll(fm.Path, ".", "-"))
			m[flagName] = confPath
		}
	}
	return m
}

// parseArgs parses --key=value and --key value pairs from CLI args.
func parseArgs(args []string) map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			continue
		}
		arg = strings.TrimPrefix(arg, "--")

		if key, val, ok := strings.Cut(arg, "="); ok {
			result[key] = val
		} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			result[arg] = args[i+1]
			i++
		} else {
			result[arg] = "true"
		}
	}
	return result
}

// Flags returns a Provider that reads --key=value pairs from CLI arguments.
func Flags[T any](args []string) fuse.Provider {
	var zero T
	return &flagsProvider{
		args:       args,
		targetType: reflect.TypeOf(zero),
	}
}
