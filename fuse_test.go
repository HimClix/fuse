package fuse_test

import (
	"strings"
	"testing"
	"time"

	"github.com/himclix/fuse"
	"github.com/himclix/fuse/parser"
	"github.com/himclix/fuse/provider"
)

// AllTypesConfig exercises: 3-level nesting, all scalar types, slices, maps, duration, secrets.
type AllTypesConfig struct {
	App struct {
		Name       string        `conf:"default:unnamed"    validate:"required"`
		Port       int           `conf:"default:8080"       validate:"required,gte=1,lte=65535"`
		Rate       float64       `validate:"gte=0,lte=1"`
		Enabled    bool          `conf:"default:false"`
		Timeout    time.Duration `conf:"default:10s"`
		MaxRetries uint          `conf:"default:3"`
		DebugPort  uint16

		Server struct {
			Host     string `validate:"required"`
			GrpcPort int    `validate:"gte=1"`

			TLS struct {
				Enabled  bool
				CertFile string `validate:"omitempty"`
				KeyFile  string
			}
		}
	}
	DB struct {
		Host            string  `conf:"env:DB_HOST"            validate:"required"`
		Port            int     `conf:"default:5432"`
		MaxOpen         int     `validate:"gte=1"`
		MaxIdle         int
		MaxLifetimeSecs float64
		ReadOnly        bool
		Password        string  `conf:"env:DB_PASSWORD,secret" validate:"required"`
	}
	Kafka struct {
		Brokers   []string `validate:"required,min=1"`
		EnableTLS bool
		GroupID   string
	}
	Tags map[string]string
}

// --- TOML ---

func TestAllTypes_TOML(t *testing.T) {
	result := loadAllTypes(t, "testdata/alltypes.toml", parser.TOML())
	assertAllTypesValues(t, result)
}

// --- YAML ---

func TestAllTypes_YAML(t *testing.T) {
	result := loadAllTypes(t, "testdata/alltypes.yaml", parser.YAML())
	assertAllTypesValues(t, result)
}

// --- JSON ---

func TestAllTypes_JSON(t *testing.T) {
	result := loadAllTypes(t, "testdata/alltypes.json", parser.JSON())
	cfg := result.Config

	// JSON doesn't have replicas/weights in the fixture, check what it does have
	if cfg.App.Name != "fuse-test" {
		t.Errorf("App.Name = %q", cfg.App.Name)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("App.Port = %d", cfg.App.Port)
	}
	if cfg.App.Rate != 0.95 {
		t.Errorf("App.Rate = %f", cfg.App.Rate)
	}
	if !cfg.App.Enabled {
		t.Error("App.Enabled should be true")
	}
	if cfg.App.Server.Host != "api.example.com" {
		t.Errorf("App.Server.Host = %q", cfg.App.Server.Host)
	}
	if cfg.App.Server.TLS.Enabled != true {
		t.Error("App.Server.TLS.Enabled should be true")
	}
	if cfg.DB.Host != "db.example.com" {
		t.Errorf("DB.Host = %q", cfg.DB.Host)
	}
	if len(cfg.Kafka.Brokers) != 2 {
		t.Errorf("Kafka.Brokers len = %d", len(cfg.Kafka.Brokers))
	}
}

// --- Override pattern ---

func TestOverrideFilePattern(t *testing.T) {
	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File("testdata/alltypes.toml", parser.TOML()),
		provider.File("testdata/override.toml", parser.TOML()),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	cfg := result.Config

	// Override changes port to 8443
	if cfg.App.Port != 8443 {
		t.Errorf("App.Port = %d, want 8443 (from override)", cfg.App.Port)
	}
	// Non-overridden values preserved from base
	if cfg.App.Name != "fuse-test" {
		t.Errorf("App.Name = %q, want fuse-test (preserved)", cfg.App.Name)
	}
	if cfg.App.Server.Host != "api.example.com" {
		t.Errorf("App.Server.Host = %q, want api.example.com (preserved)", cfg.App.Server.Host)
	}
}

// --- Env vars ---

func TestEnvVarOverridesFile(t *testing.T) {
	t.Setenv("FTEST_APP_PORT", "3000")
	t.Setenv("DB_HOST", "env-db.internal")
	t.Setenv("DB_PASSWORD", "env-pass-xyz")

	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File("testdata/alltypes.toml", parser.TOML()),
		provider.Env[AllTypesConfig]("FTEST"),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	cfg := result.Config

	if cfg.App.Port != 3000 {
		t.Errorf("App.Port = %d, want 3000 (env override)", cfg.App.Port)
	}
	if cfg.DB.Host != "env-db.internal" {
		t.Errorf("DB.Host = %q, want env-db.internal (explicit env tag)", cfg.DB.Host)
	}
	if cfg.DB.Password != "env-pass-xyz" {
		t.Errorf("DB.Password = %q, want env-pass-xyz", cfg.DB.Password)
	}
}

// --- Flags ---

func TestFlagsOverride(t *testing.T) {
	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File("testdata/alltypes.toml", parser.TOML()),
		provider.Flags[AllTypesConfig]([]string{"--app-port=4000", "--db-host=flag-db.internal"}),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if result.Config.App.Port != 4000 {
		t.Errorf("App.Port = %d, want 4000 (flag override)", result.Config.App.Port)
	}
}

// --- DotEnv ---

func TestDotEnvProvider(t *testing.T) {
	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File("testdata/alltypes.toml", parser.TOML()),
		provider.DotEnv("testdata/test.env"),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// .env sets DB_HOST, DB_PASSWORD, APP_PORT as flat keys
	// These are flat (no struct awareness), so they land in the map as top-level keys
	_ = result
}

// --- Defaults only ---

// DefaultsOnlyConfig has only fields that can be satisfied by defaults + env.
type DefaultsOnlyConfig struct {
	App struct {
		Port    int           `conf:"default:8080"  validate:"required"`
		Host    string        `conf:"default:localhost"`
		Env     string        `conf:"default:dev"`
		Timeout time.Duration `conf:"default:10s"`
		Retries uint          `conf:"default:3"`
	}
	DB struct {
		Host     string `conf:"env:DB_HOST"            validate:"required"`
		Port     int    `conf:"default:5432"`
		Password string `conf:"env:DB_PASSWORD,secret" validate:"required"`
	}
}

func TestDefaultsOnly(t *testing.T) {
	t.Setenv("DB_HOST", "defaults-db")
	t.Setenv("DB_PASSWORD", "defaults-pass")

	result, err := fuse.Load[DefaultsOnlyConfig](
		provider.Defaults[DefaultsOnlyConfig](),
		provider.Env[DefaultsOnlyConfig](""),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	cfg := result.Config

	if cfg.App.Port != 8080 {
		t.Errorf("App.Port = %d, want 8080 (default)", cfg.App.Port)
	}
	if cfg.App.Host != "localhost" {
		t.Errorf("App.Host = %q, want localhost (default)", cfg.App.Host)
	}
	if cfg.App.Timeout != 10*time.Second {
		t.Errorf("App.Timeout = %v, want 10s (default)", cfg.App.Timeout)
	}
	if cfg.App.Retries != 3 {
		t.Errorf("App.Retries = %d, want 3 (default)", cfg.App.Retries)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("DB.Port = %d, want 5432 (default)", cfg.DB.Port)
	}
	if cfg.DB.Host != "defaults-db" {
		t.Errorf("DB.Host = %q, want defaults-db (env)", cfg.DB.Host)
	}
}

// --- Validation failures ---

func TestValidationCollectsAllErrors(t *testing.T) {
	_, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
	)
	if err == nil {
		t.Fatal("expected validation error")
	}

	verrs, ok := err.(fuse.ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}

	// Should have multiple errors: DB.Host, DB.Password, App.Server.Host, Kafka.Brokers, DB.MaxOpen
	if len(verrs) < 3 {
		t.Errorf("expected at least 3 validation errors, got %d:\n%v", len(verrs), err)
	}

	// Check specific errors exist
	tags := make(map[string]bool)
	for _, ve := range verrs {
		tags[ve.Path+":"+ve.Tag] = true
	}
	for _, want := range []string{"DB.Host:required", "DB.Password:required"} {
		if !tags[want] {
			t.Errorf("missing expected error: %s", want)
		}
	}
}

// --- Invalid file ---

func TestInvalidFileErrors(t *testing.T) {
	_, err := fuse.Load[AllTypesConfig](
		provider.File("testdata/invalid.toml", parser.TOML()),
	)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
	ce, ok := err.(*fuse.ConfigError)
	if !ok {
		t.Fatalf("expected ConfigError, got %T", err)
	}
	if ce.Provider != "testdata/invalid.toml" {
		t.Errorf("Provider = %q", ce.Provider)
	}
}

func TestMissingFileErrors(t *testing.T) {
	_, err := fuse.Load[AllTypesConfig](
		provider.File("testdata/nonexistent.toml", parser.TOML()),
	)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestOptionalFileMissing(t *testing.T) {
	t.Setenv("DB_HOST", "test-db")
	t.Setenv("DB_PASSWORD", "test-pass")

	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File("testdata/alltypes.toml", parser.TOML()),
		provider.OptionalFile("testdata/nonexistent.toml", parser.TOML()),
		provider.Env[AllTypesConfig](""),
	)
	if err != nil {
		t.Fatalf("OptionalFile should not fail: %v", err)
	}
	if result.Config.App.Port != 8080 {
		t.Errorf("App.Port = %d", result.Config.App.Port)
	}
}

// --- Source tracking ---

func TestSourceTracking(t *testing.T) {
	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File("testdata/alltypes.toml", parser.TOML()),
		provider.File("testdata/override.toml", parser.TOML()),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if src := result.Sources["app.port"]; src != "testdata/override.toml" {
		t.Errorf("source[app.port] = %q, want testdata/override.toml", src)
	}
	if src := result.Sources["app.name"]; src != "testdata/alltypes.toml" {
		t.Errorf("source[app.name] = %q, want testdata/alltypes.toml", src)
	}
}

// --- Summary ---

func TestSummaryMasksSecrets(t *testing.T) {
	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File("testdata/alltypes.toml", parser.TOML()),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	summary := result.Summary()
	if !strings.Contains(summary, "*****") {
		t.Error("summary should mask secret fields")
	}
	if strings.Contains(summary, "db-secret") {
		t.Error("summary must NOT contain actual secret value")
	}
	if !strings.Contains(summary, "testdata/alltypes.toml") {
		t.Error("summary should show file source")
	}
}

// --- MustLoad ---

func TestMustLoadPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoad should panic on validation error")
		}
	}()
	fuse.MustLoad[AllTypesConfig](provider.Defaults[AllTypesConfig]())
}

// --- Full precedence chain ---

func TestFullPrecedenceChain(t *testing.T) {
	t.Setenv("PREC_APP_PORT", "5000")

	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),                        // port=8080
		provider.File("testdata/alltypes.toml", parser.TOML()),     // port=8080
		provider.File("testdata/override.toml", parser.TOML()),     // port=8443
		provider.Env[AllTypesConfig]("PREC"),                       // port=5000
		provider.Flags[AllTypesConfig]([]string{"--app-port=6000"}),// port=6000
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if result.Config.App.Port != 6000 {
		t.Errorf("App.Port = %d, want 6000 (flags wins last)", result.Config.App.Port)
	}
}

// --- Helpers ---

func loadAllTypes(t *testing.T, path string, p fuse.Parser) *fuse.Result[AllTypesConfig] {
	t.Helper()
	result, err := fuse.Load[AllTypesConfig](
		provider.Defaults[AllTypesConfig](),
		provider.File(path, p),
	)
	if err != nil {
		t.Fatalf("Load(%s) failed: %v", path, err)
	}
	return result
}

func assertAllTypesValues(t *testing.T, result *fuse.Result[AllTypesConfig]) {
	t.Helper()
	cfg := result.Config

	// Scalars
	if cfg.App.Name != "fuse-test" {
		t.Errorf("App.Name = %q", cfg.App.Name)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("App.Port = %d", cfg.App.Port)
	}
	if cfg.App.Rate != 0.95 {
		t.Errorf("App.Rate = %f", cfg.App.Rate)
	}
	if !cfg.App.Enabled {
		t.Error("App.Enabled should be true")
	}
	if cfg.App.Timeout != 5*time.Second {
		t.Errorf("App.Timeout = %v, want 5s", cfg.App.Timeout)
	}
	if cfg.App.MaxRetries != 3 {
		t.Errorf("App.MaxRetries = %d", cfg.App.MaxRetries)
	}
	if cfg.App.DebugPort != 6060 {
		t.Errorf("App.DebugPort = %d", cfg.App.DebugPort)
	}

	// 2-level nesting
	if cfg.App.Server.Host != "api.example.com" {
		t.Errorf("App.Server.Host = %q", cfg.App.Server.Host)
	}
	if cfg.App.Server.GrpcPort != 9090 {
		t.Errorf("App.Server.GrpcPort = %d", cfg.App.Server.GrpcPort)
	}

	// 3-level nesting
	if !cfg.App.Server.TLS.Enabled {
		t.Error("App.Server.TLS.Enabled should be true")
	}
	if cfg.App.Server.TLS.CertFile != "/etc/ssl/cert.pem" {
		t.Errorf("App.Server.TLS.CertFile = %q", cfg.App.Server.TLS.CertFile)
	}

	// Float
	if cfg.DB.MaxLifetimeSecs != 300.5 {
		t.Errorf("DB.MaxLifetimeSecs = %f", cfg.DB.MaxLifetimeSecs)
	}

	// Bool false
	if cfg.DB.ReadOnly {
		t.Error("DB.ReadOnly should be false")
	}

	// Slice of strings
	if len(cfg.Kafka.Brokers) != 2 || cfg.Kafka.Brokers[0] != "broker1:9092" {
		t.Errorf("Kafka.Brokers = %v", cfg.Kafka.Brokers)
	}

	// Map[string]string
	if cfg.Tags["env"] != "prod" {
		t.Errorf("Tags[env] = %q", cfg.Tags["env"])
	}
	if cfg.Tags["team"] != "platform" {
		t.Errorf("Tags[team] = %q", cfg.Tags["team"])
	}
}

// --- Boot() convenience function ---

type BootConfig struct {
	App struct {
		Port int    `conf:"default:8080"     validate:"required"`
		Env  string `conf:"default:dev"      validate:"required"`
	}
	DB struct {
		Host     string `conf:"default:localhost" validate:"required"`
		Password string `conf:"env:DB_PASSWORD"   validate:"required"`
	}
}

func TestBootLoadsTOML(t *testing.T) {
	t.Setenv("DB_PASSWORD", "boot-pass")

	result, err := fuse.Boot[BootConfig](fuse.BootOpts{
		ConfigDir: "./testdata/bootconfig",
		Env:       "prod",
		Format:    "toml",
	})
	if err != nil {
		t.Fatalf("Boot failed: %v", err)
	}
	if result.Config.App.Port != 9090 {
		t.Errorf("App.Port = %d, want 9090", result.Config.App.Port)
	}
	if result.Config.App.Env != "prod" {
		t.Errorf("App.Env = %q, want prod", result.Config.App.Env)
	}
}

func TestBootDefaultEnvIsDev(t *testing.T) {
	t.Setenv("DB_PASSWORD", "boot-pass")

	result, err := fuse.Boot[BootConfig](fuse.BootOpts{
		ConfigDir: "./testdata/bootconfig",
	})
	if err != nil {
		t.Fatalf("Boot failed: %v", err)
	}
	if result.Config.App.Env != "dev" {
		t.Errorf("App.Env = %q, want dev (default)", result.Config.App.Env)
	}
}

func TestBootEnvVarOverridesFile(t *testing.T) {
	t.Setenv("DB_PASSWORD", "from-env")

	result, err := fuse.Boot[BootConfig](fuse.BootOpts{
		ConfigDir: "./testdata/bootconfig",
		Env:       "prod",
	})
	if err != nil {
		t.Fatalf("Boot failed: %v", err)
	}
	if result.Config.DB.Password != "from-env" {
		t.Errorf("DB.Password = %q, want from-env", result.Config.DB.Password)
	}
}

func TestBootValidationFails(t *testing.T) {
	_, err := fuse.Boot[BootConfig](fuse.BootOpts{
		ConfigDir: "./testdata/bootconfig",
	})
	if err == nil {
		t.Fatal("expected validation error (DB.Password missing)")
	}
}

func TestBootUnsupportedFormat(t *testing.T) {
	_, err := fuse.Boot[BootConfig](fuse.BootOpts{
		ConfigDir: "./testdata/bootconfig",
		Format:    "xml",
	})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestBootMissingConfigDir(t *testing.T) {
	t.Setenv("DB_PASSWORD", "x")

	_, err := fuse.Boot[BootConfig](fuse.BootOpts{
		ConfigDir: "./nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for missing config dir")
	}
}

func TestMustBootPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBoot should panic on error")
		}
	}()
	fuse.MustBoot[BootConfig](fuse.BootOpts{ConfigDir: "./nonexistent"})
}
