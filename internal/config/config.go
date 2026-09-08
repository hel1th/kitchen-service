package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Log      LogConfig      `env-prefix:"LOG_"`
	HTTP     HTTPConfig     `env-prefix:"HTTP_"`
	Postgres PostgresConfig `env-prefix:"POSTGRES_"`
}

type LogConfig struct {
	Level string `env:"LEVEL" env-default:"info"`
}

type HTTPConfig struct {
	Host            string `env:"HOST"              env-default:""`
	Port            string `env:"PORT"              env-default:"8080"`
	APIRouterPrefix string `env:"API_ROUTER_PREFIX" env-default:"/api/v1"`
}

type PostgresConfig struct {
	User     string `env:"USER"     env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
	Host     string `env:"HOST"     env-default:"localhost"`
	Port     string `env:"PORT"     env-default:"5432"`
	DB       string `env:"DB"       env-required:"true"`
	SSLMode  string `env:"SSLMODE"  env-default:"disable"`
}

// DSN returns a connection string for PostgreSQL based on config
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode,
	)
}

func Load() (*Config, error) {
	var cfg Config
	// Attempt to load from .env file if it exists
	_ = cleanenv.ReadConfig(".env", &cfg)

	// ReadEnv will overwrite with actual environment variables
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
