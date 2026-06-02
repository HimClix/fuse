package fuse

import (
	"fmt"
	"os"
	"reflect"

	"github.com/himclix/fuse/internal/decode"
	"github.com/himclix/fuse/internal/merge"
	"github.com/himclix/fuse/internal/resolve"
	"github.com/himclix/fuse/internal/summary"
	"github.com/himclix/fuse/validate"
)

// Load loads configuration from multiple providers into a typed struct.
// Providers are applied in order — later providers override earlier ones.
func Load[T any](providers ...Provider) (*Result[T], error) {
	var results []merge.Result
	for _, p := range providers {
		data, err := p.Load()
		if err != nil {
			return nil, &ConfigError{Provider: p.Name(), Err: err}
		}
		if data != nil {
			results = append(results, merge.Result{Name: p.Name(), Data: data})
		}
	}

	merged, sources := merge.All(results)

	var cfg T
	if err := decode.Map(merged, &cfg); err != nil {
		return nil, &ConfigError{Err: fmt.Errorf("decode: %w", err)}
	}

	engine := validate.New()
	if verrs := engine.Validate(&cfg); verrs != nil {
		return nil, verrs
	}

	return &Result[T]{Config: &cfg, Sources: sources}, nil
}

// MustLoad is like Load but panics on error.
func MustLoad[T any](providers ...Provider) *Result[T] {
	r, err := Load[T](providers...)
	if err != nil {
		panic(fmt.Sprintf("fuse.MustLoad: %v", err))
	}
	return r
}

// BootOpts configures the Boot convenience loader.
type BootOpts struct {
	ConfigDir string // Default: "./config"
	Env       string // Default: "dev"
	EnvPrefix string // Prefix for auto-derived env var names
	Format    string // Default: "toml"
}

// Boot loads config following the standard service boot pattern:
//  1. Read {ConfigDir}/default.{format}
//  2. Read {ConfigDir}/{env}.{format} (optional)
//  3. Read environment variables
//  4. Validate + unmarshal into *T
func Boot[T any](opts BootOpts) (*Result[T], error) {
	if opts.ConfigDir == "" {
		opts.ConfigDir = "./config"
	}
	if opts.Env == "" {
		opts.Env = "dev"
	}
	if opts.Format == "" {
		opts.Format = "toml"
	}

	var zero T
	t := reflect.TypeOf(zero)

	providers := []Provider{
		asProvider(resolve.Defaults(t)),
		asProvider(resolve.File(opts.ConfigDir+"/default."+opts.Format, opts.Format, false)),
		asProvider(resolve.File(opts.ConfigDir+"/"+opts.Env+"."+opts.Format, opts.Format, true)),
		asProvider(resolve.EnvVars(t, opts.EnvPrefix)),
	}

	return Load[T](providers...)
}

// MustBoot is like Boot but panics on error.
func MustBoot[T any](opts BootOpts) *Result[T] {
	r, err := Boot[T](opts)
	if err != nil {
		panic(fmt.Sprintf("fuse.MustBoot: %v", err))
	}
	return r
}

// Summary returns a formatted config summary string.
func (r *Result[T]) Summary() string {
	return summary.Format(r.Config, r.Sources)
}

// Print writes the config summary to stdout.
func (r *Result[T]) Print() {
	summary.Write(os.Stdout, r.Config, r.Sources)
}

// asProvider wraps an internal resolve.Provider as a fuse.Provider.
type internalProvider struct{ inner resolve.Provider }

func (p *internalProvider) Name() string                  { return p.inner.Name() }
func (p *internalProvider) Load() (map[string]any, error) { return p.inner.Load() }

func asProvider(p resolve.Provider) Provider { return &internalProvider{inner: p} }
