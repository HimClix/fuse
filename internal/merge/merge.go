package merge

import "strings"

// Result holds one provider's output.
type Result struct {
	Name string
	Data map[string]any
}

// All merges multiple provider results in order (later wins).
func All(results []Result) (map[string]any, map[string]string) {
	merged := make(map[string]any)
	sources := make(map[string]string)

	for _, pr := range results {
		deep(merged, pr.Data, sources, pr.Name, "")
	}
	return merged, sources
}

func deep(dst, src map[string]any, sources map[string]string, provider, prefix string) {
	for key, srcVal := range src {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		dstVal, exists := dst[key]
		if !exists {
			dst[key] = srcVal
			trackSources(sources, srcVal, provider, fullKey)
			continue
		}

		srcMap, srcIsMap := srcVal.(map[string]any)
		dstMap, dstIsMap := dstVal.(map[string]any)
		if srcIsMap && dstIsMap {
			deep(dstMap, srcMap, sources, provider, fullKey)
			continue
		}

		dst[key] = srcVal
		trackSources(sources, srcVal, provider, fullKey)
	}
}

func trackSources(sources map[string]string, val any, provider, fullKey string) {
	sources[fullKey] = provider
	if m, ok := val.(map[string]any); ok {
		for k, v := range m {
			trackSources(sources, v, provider, fullKey+"."+k)
		}
	}
}

// LookupNested retrieves a value from a nested map by dot-separated path.
func LookupNested(m map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = m
	for _, part := range parts {
		cm, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = cm[part]
		if !ok {
			found := false
			for k, v := range cm {
				if strings.EqualFold(k, part) {
					current = v
					found = true
					break
				}
			}
			if !found {
				return nil, false
			}
		}
	}
	return current, true
}

// SetNested sets a value in a nested map by dot-separated path.
func SetNested(m map[string]any, path string, val any) {
	parts := strings.Split(path, ".")
	current := m
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = val
			return
		}
		sub, ok := current[part].(map[string]any)
		if !ok {
			sub = make(map[string]any)
			current[part] = sub
		}
		current = sub
	}
}
