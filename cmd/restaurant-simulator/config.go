package main

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port              string        `env:"PORT" env-default:"8081"`
	KitchenServiceURL string        `env:"KITCHEN_SERVICE_URL" env-default:"http://kitchen-service:8080"`
	CookingDelay      time.Duration `env:"COOKING_DELAY" env-default:"3s"`
	ReadyDelay        time.Duration `env:"READY_DELAY" env-default:"5s"`
	CallbackTimeout   time.Duration `env:"CALLBACK_TIMEOUT" env-default:"2s"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	// Ignore error as .env file might not exist
	_ = cleanenv.ReadConfig(".env", &cfg)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
