# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-02

### Added
- `Load[T]()` and `MustLoad[T]()` generics API
- `Result[T]` with `.Summary()` and `.Print()` for startup observability
- **Providers**: defaults, file, optional file, auto-detect file, env vars, CLI flags, .env
- **Parsers**: TOML, YAML, JSON (behind `Parser` interface)
- **53 built-in validation rules**: compare, string, format, cross-field, collection, filesystem
- Custom validator registration via `validate.Engine.Register()`
- Deep merge with override file pattern (base + diffs only)
- Source tracking — know which provider set each field
- Secret masking via `conf:"secret"` tag
- `conf:"default:V,env:NAME,flag:NAME,secret"` struct tags
- `validate:"required,min=1,oneof=a b c"` struct tags (go-playground/validator compatible syntax)
