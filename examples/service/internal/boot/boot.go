package boot

import (
	"fmt"
	"os"

	"github.com/himclix/fuse"
	"github.com/himclix/fuse/examples/service/internal/config"
	"github.com/himclix/fuse/parser"
	"github.com/himclix/fuse/provider"
	"github.com/himclix/fuse/validate"
)

// Result holds the booted config and fuse metadata.
var Result *fuse.Result[config.Config]

// Init loads config and initializes all core dependencies.
//
// This follows the standard Go service boot pattern:
//  1. Read config/default.toml         — base config with all keys
//  2. Read config/{env}.toml            — env-specific overrides (only diffs)
//  3. Read environment variables       — runtime overrides
//  4. Validate + unmarshal into struct — fail fast on bad config
//
// The simplest way (one-liner):
//
//	result, err := fuse.Boot[config.Config](fuse.BootOpts{
//	    ConfigDir: "./config",
//	    Env:       GetEnv(),
//	    EnvPrefix: "APP",
//	})
//
// For full control over the provider chain, use fuse.Load[T]() instead (shown below).
func Init() error {
	env := GetEnv()

	// Register custom validators
	engine := validate.New()
	engine.Register("is_port", func(info validate.FieldInfo) bool {
		v, ok := info.Value.(int)
		return ok && v >= 1 && v <= 65535
	})

	var err error
	Result, err = fuse.Load[config.Config](
		provider.Defaults[config.Config](),
		provider.File("config/default.toml", parser.TOML()),
		provider.OptionalFile("config/"+env+".toml", parser.TOML()),
		provider.Env[config.Config](""),
	)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	Result.Print()
	return nil
}

// Cfg returns the loaded config. Panics if Init() was not called.
func Cfg() *config.Config {
	return Result.Config
}

func GetEnv() string {
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}
	return "dev"
}
