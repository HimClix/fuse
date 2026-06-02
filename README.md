# fuse

Configuration management for Go that finally makes sense.

---

## The Problem

You've written this code before. Every service, every time:

```go
viper.SetConfigName("default")
viper.SetConfigType("toml")
viper.AddConfigPath("./config")
viper.AutomaticEnv()
viper.ReadInConfig()
viper.Unmarshal(&config)
```

It works. Until it doesn't.

The port is `0` because someone forgot to set an env var. The password leaks into your startup logs. You spend an hour debugging why config values are wrong, only to discover Viper silently lowercased your TOML keys. You have 12 copies of a 500-line config file because each environment needs its own `prod.toml`, `stage.toml`, `perf.toml` — and they're 95% identical.

**fuse** is the library you wish existed when you started that service.

---

## What It Does

One struct. One function call. Everything loaded, merged, validated, and observable.

```go
result, err := fuse.Boot[Config](fuse.BootOpts{
    ConfigDir: "./config",
    Env:       os.Getenv("APP_ENV"),
    EnvPrefix: "APP",
})
```

That single call:

1. Reads `config/default.toml` — your base config with every key
2. Reads `config/prod.toml` — only the values that differ (5 lines, not 500)
3. Reads environment variables — runtime overrides
4. Decodes into your typed struct — no `interface{}`, no type assertions
5. Validates every field — `required`, `email`, `gte=1`, `oneof=dev stage prod`
6. Tracks which provider set each value — for debugging
7. Returns an immutable result — no race conditions, no surprises

If anything is wrong, you get every problem at once:

```
fuse: 3 validation error(s):
  - field "DB.Password" failed on "required" (value=)
  - field "App.Port" failed on "gte" (param="1", value=0)
  - field "Kafka.Brokers" failed on "min" (param="1", value=[])
```

Your service crashes at startup, not at 3am when a zero-value port silently breaks something.

---

## Install

```bash
go get github.com/himclix/fuse
```

Requires Go 1.22+ (generics).

---

## Quick Start

### Define your config

```go
// internal/config/config.go

type Config struct {
    App struct {
        Env  string `conf:"default:dev"     validate:"required,oneof=dev stage prod"`
        Port int    `conf:"default:8080"    validate:"required,gte=1,lte=65535"`
    }
    DB struct {
        Host     string `conf:"default:localhost" validate:"required,hostname"`
        Password string `conf:"env:DB_PASSWORD,secret" validate:"required,min=8"`
    }
    Kafka struct {
        Brokers []string `validate:"required,min=1,dive,hostname_port"`
    }
}
```

Two tag namespaces, each with a clear job:

- **`conf`** — where does this value come from? (`default`, `env`, `flag`, `secret`)
- **`validate`** — what makes this value valid? (`required`, `email`, `gte=1`, `oneof=...`)

### Load it

The simplest way — mirrors the standard service boot pattern:

```go
result, err := fuse.Boot[Config](fuse.BootOpts{
    ConfigDir: "./config",
    Env:       os.Getenv("APP_ENV"),
    EnvPrefix: "APP",
})
if err != nil {
    log.Fatal(err)
}
result.Print()
cfg := result.Config
```

When you need full control over the provider chain:

```go
import (
    "github.com/himclix/fuse"
    "github.com/himclix/fuse/parser"
    "github.com/himclix/fuse/provider"
)

result, err := fuse.Load[Config](
    provider.Defaults[Config](),                              // struct tag defaults
    provider.File("config/default.toml", parser.TOML()),      // base config
    provider.OptionalFile("config/prod.toml", parser.TOML()), // diffs only
    provider.Env[Config]("APP"),                              // APP_DB_HOST, etc.
    provider.Flags[Config](os.Args[1:]),                      // --db-host=x
)
```

### See what you loaded

```go
result.Print()
```

```
  KEY          VALUE               SOURCE
  ─────────   ──────────────────  ──────────────────────────
  app.port     8443                config/prod.toml
  app.env      prod                env
  db.host      prod-db.internal    config/prod.toml
  db.password  *****               env
  kafka.brokers [b1:9092 b2:9092] config/prod.toml
```

Every value. Where it came from. Secrets masked. At a glance.

---

## The Override Pattern

This is the pattern that eliminates config file duplication.

Instead of maintaining 12 nearly-identical 500-line config files — one per environment — you maintain one complete `default.toml` and tiny override files that contain only what differs:

```
config/
├── default.toml           ← 500 lines, every key
├── prod.toml     ← 8 lines, only what differs
├── stage.override.toml    ← 5 lines
└── dev.override.toml      ← 3 lines (or nothing at all)
```

fuse deep-merges them. Nested maps merge recursively. Scalars overwrite. Everything else is preserved from the base.

---

## Providers

Providers are where your config comes from. They execute in the order you specify — later providers override earlier ones.

| Provider | What it does |
|---|---|
| `provider.Defaults[T]()` | Reads `conf:"default:..."` from struct tags |
| `provider.File(path, parser)` | Reads and parses a config file |
| `provider.OptionalFile(path, parser)` | Same, but no error if the file is missing |
| `provider.Env[T](prefix)` | Reads environment variables |
| `provider.DotEnv(path)` | Reads a `.env` file |
| `provider.Flags[T](args)` | Reads `--key=value` from CLI arguments |
| `provider.AutoFile(path, parsers)` | Auto-detects format from file extension |

The standard precedence — `defaults → file → env → flags` — matches the twelve-factor pattern that most teams already use.

### Custom providers

Any type that implements `Provider` works:

```go
type Provider interface {
    Name() string
    Load() (map[string]any, error)
}
```

Two methods. That's it. Build providers for Vault, Consul, S3, a database — whatever your team needs.

## Parsers

| Parser | Format | Dependency |
|---|---|---|
| `parser.TOML()` | TOML | `pelletier/go-toml/v2` |
| `parser.YAML()` | YAML | `gopkg.in/yaml.v3` |
| `parser.JSON()` | JSON | stdlib |

Any type that implements `Parser` works:

```go
type Parser interface {
    Parse(data []byte) (map[string]any, error)
}
```

---

## Validation

54 built-in rules, zero external dependencies. The syntax is compatible with [go-playground/validator](https://github.com/go-playground/validator) so you already know it.

### Core
`required`, `omitempty`

### Comparisons
For numbers, compares the value. For strings, slices, and maps, compares the length.

`eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `min`, `max`, `len`, `oneof`

### Strings
`alpha`, `alphanum`, `numeric`, `lowercase`, `uppercase`, `contains`, `excludes`, `startswith`, `endswith`

### Formats
`email`, `url`, `http_url`, `uri`, `ip`, `ipv4`, `ipv6`, `cidr`, `mac`, `hostname`, `fqdn`, `hostname_port`, `uuid`, `uuid4`, `json`, `jwt`, `base64`, `semver`, `datetime`, `timezone`, `cron`

### Filesystem
`dir`, `dirpath`, `file`, `filepath`

### Cross-field
When one field's validity depends on another:

`required_if`, `required_unless`, `required_with`, `required_without`, `eqfield`, `nefield`, `gtfield`, `ltfield`

```go
Replicas int `validate:"required_if=Env prod,gte=1"`
```

Replicas is required only when Env is "prod". In dev, zero replicas is fine.

### Collections
`unique`, `dive`

```go
Brokers []string `validate:"required,min=1,dive,hostname_port"`
```

The slice must be non-empty, and each element must be a valid `host:port`.

### Combinators
- Comma = AND: `validate:"required,min=1,max=100"`
- Pipe = OR: `validate:"email|url"`

### Custom validators

```go
engine := validate.New()
engine.Register("is_port", func(info validate.FieldInfo) bool {
    v, ok := info.Value.(int)
    return ok && v >= 1 && v <= 65535
})
```

Same interface as built-in validators. Register once, use everywhere.

---

## Config Tags

### `conf` — Loading directives

| Tag | Example | What it does |
|---|---|---|
| `default:V` | `conf:"default:8080"` | Default value if no provider sets it |
| `env:NAME` | `conf:"env:DB_HOST"` | Explicit env var name (skips prefix) |
| `flag:NAME` | `conf:"flag:db-host"` | Explicit CLI flag name |
| `secret` | `conf:"secret"` | Masked as `*****` in summary output |

### `validate` — Validation rules

See the [Validation](#validation) section above for all 54 rules.

### Together

```go
type Config struct {
    Port     int      `conf:"default:8080,env:APP_PORT"  validate:"required,gte=1,lte=65535"`
    DBPass   string   `conf:"env:DB_PASS,secret"         validate:"required,min=8"`
    LogLevel string   `conf:"default:info"                validate:"oneof=debug info warn error"`
    Brokers  []string `conf:"env:KAFKA_BROKERS"           validate:"required,min=1,dive,hostname_port"`
}
```

---

## Source Tracking

Every value knows where it came from:

```go
result, _ := fuse.Load[Config](...)

// Programmatic inspection
src := result.Sources["db.password"]  // → "env"
src = result.Sources["app.port"]      // → "config/prod.toml"
```

This is the feature you'll use at 2am when something is wrong and you need to know *which* provider set that unexpected value.

---

## Project Layout

When you use fuse in a real service, here's how the pieces fit together:

```
my-service/
├── cmd/api/main.go                  ← calls boot.Init(), starts server
├── internal/
│   ├── boot/boot.go                 ← fuse.Load[Config](...), provider chain
│   └── config/config.go             ← Config struct with conf + validate tags
├── config/
│   ├── default.toml                 ← full base config
│   └── prod.toml           ← only values that differ
└── go.mod
```

See the [examples/service](examples/service) directory for a complete working example.

---

## Architecture

```
fuse (root)         → Load[T], Boot[T], Result[T], interfaces, merge, decode
├── parser/         → TOML, YAML, JSON (implements fuse.Parser)
├── provider/       → defaults, file, env, flags, dotenv (implements fuse.Provider)
└── validate/       → 54 built-in rules, custom registration (standalone, no fuse dependency)
```

No circular dependencies. The `validate/` package is fully standalone — you can use it independently of fuse if you want.

---

## Comparison

| | fuse | Viper | koanf | envconfig |
|---|---|---|---|---|
| **API** | `Load[T]()` — generics | `Get()` → `interface{}` | `Unmarshal()` | `Process()` |
| **Validation** | 54 built-in rules | None | None | `required` only |
| **Secret masking** | `conf:"secret"` | No | No | No |
| **Source tracking** | Per field | No | No | No |
| **Startup summary** | `result.Print()` | No | No | No |
| **Override files** | Deep merge built-in | Manual | Manual | N/A |
| **Key casing** | Preserved | Force-lowercased | Preserved | N/A |
| **Dependencies** | 2 | 15+ | 3+ | 0 |

---

## License

MIT
