package config

// Config is the top-level application configuration.
// Each section maps to a [section] in the TOML file.
type Config struct {
	App      App
	Database Database
	Kafka    Kafka
	Auth     Auth
	Sentry   Sentry
}

type App struct {
	Env             string `conf:"default:dev"          validate:"required,oneof=dev stage prod"`
	ServiceName     string `conf:"default:my-service"   validate:"required"`
	Port            string `conf:"default:8080"         validate:"required"`
	MetricsPort     string `conf:"default:8002"`
	ShutdownTimeout int    `conf:"default:5"            validate:"gte=1"`
	ShutdownDelay   int    `conf:"default:2"            validate:"gte=0"`
	GitCommitHash   string `conf:"default:unknown"`
}

type Database struct {
	Host     string `conf:"default:localhost"              validate:"required"`
	Port     int    `conf:"default:5432"                   validate:"gte=1,lte=65535"`
	Username string `conf:"default:root"                   validate:"required"`
	Password string `conf:"env:DB_PASSWORD,secret"         validate:"required"`
	Name     string `conf:"default:myapp"                  validate:"required"`
	SSLMode  string `conf:"default:disable"                validate:"oneof=disable require verify-full"`
	Pool     Pool
}

type Pool struct {
	MaxOpenConnections    int `conf:"default:10"  validate:"gte=1"`
	MaxIdleConnections    int `conf:"default:5"   validate:"gte=0"`
	ConnectionMaxLifetime int `conf:"default:300" validate:"gte=0"`
}

type Kafka struct {
	Brokers   []string `validate:"required,min=1,dive,hostname_port"`
	EnableTLS bool     `conf:"default:false"`
	GroupID   string   `conf:"default:my-consumer"  validate:"required"`
}

type Auth struct {
	Token          string `conf:"env:AUTH_TOKEN,secret"   validate:"required"`
	TokenHeaderKey string `conf:"default:X-Auth-Key"      validate:"required"`
}

type Sentry struct {
	DSN     string `conf:"secret"`
	Enabled bool   `conf:"default:false"`
}
