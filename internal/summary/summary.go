package summary

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/himclix/fuse/internal/tag"
)

// Write prints a formatted config summary to w.
func Write(w io.Writer, cfg any, sources map[string]string) {
	rv := reflect.ValueOf(cfg)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	meta := tag.Get(rv.Type())
	rows := collectRows(rv, meta, sources)

	if len(rows) == 0 {
		fmt.Fprintln(w, "(no config fields)")
		return
	}

	maxKey, maxVal, maxSrc := 4, 5, 6
	for _, row := range rows {
		if len(row.key) > maxKey {
			maxKey = len(row.key)
		}
		if len(row.value) > maxVal {
			maxVal = len(row.value)
		}
		if len(row.source) > maxSrc {
			maxSrc = len(row.source)
		}
	}

	format := fmt.Sprintf("  %%-%ds  %%-%ds  %%s\n", maxKey, maxVal)
	fmt.Fprintf(w, format, "KEY", "VALUE", "SOURCE")
	fmt.Fprintf(w, "  %s  %s  %s\n", strings.Repeat("─", maxKey), strings.Repeat("─", maxVal), strings.Repeat("─", maxSrc))

	for _, row := range rows {
		fmt.Fprintf(w, format, row.key, row.value, row.source)
	}
}

// Format returns the summary as a string.
func Format(cfg any, sources map[string]string) string {
	var b strings.Builder
	Write(&b, cfg, sources)
	return b.String()
}

type row struct {
	key, value, source string
}

func collectRows(rv reflect.Value, meta *tag.StructMeta, sources map[string]string) []row {
	var rows []row

	for _, fm := range meta.Fields {
		fv := fieldByIndex(rv, fm.Index)
		if !fv.IsValid() {
			continue
		}

		displayVal := fmt.Sprintf("%v", fv.Interface())
		if fm.Conf.Secret && displayVal != "" && displayVal != "0" && displayVal != "false" {
			displayVal = "*****"
		}

		source := "(not set)"
		confKey := tag.ConfigPath(fm.Path)
		if s, ok := sources[confKey]; ok {
			source = s
		}

		rows = append(rows, row{key: confKey, value: displayVal, source: source})
	}
	return rows
}

func fieldByIndex(rv reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return reflect.Value{}
			}
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct || i >= rv.NumField() {
			return reflect.Value{}
		}
		rv = rv.Field(i)
	}
	return rv
}
