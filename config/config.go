package config

import (
	"fmt"
	"path"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		HTTP     `yaml:"http"`
		Database `yaml:"database"`
		Log      `yaml:"log"`
	}

	HTTP struct {
		Port        string        `env-required:"true" yaml:"port" env:"SERVER_PORT"`
		Address     string        `env-required:"true" yaml:"address" env:"SERVER_ADDRESS"`
		Timeout     string        `env-required:"true" yaml:"timeout"`
		IdleTimeout time.Duration `env-required:"true" yaml:"idle_timeout"`
	}

	Database struct {
		Conn        string `env-required:"true" env:"POSTGRES_CONN"`
		MaxPoolSize int    `env-required:"true" yaml:"max_pool_size" env:"MAX_POOL_SIZE"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"level" env:"LOG_LEVEL"`
	}
)

func NewConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(path.Join("./", configPath), cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = cleanenv.UpdateEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("error updating env: %w", err)
	}

	return cfg, nil
}
