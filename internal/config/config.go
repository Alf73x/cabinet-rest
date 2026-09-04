package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/numbergroup/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
	// StoragePath string  `yaml:"storage_path" env-required:"true"`
	StoragePath string
	HTTPServer  `yaml:"http_server"`
	JWT         JWT `yaml:"jwt"`
}

type JWT struct {
	Secret string        `yaml:"secret"`
	TTL    time.Duration `yaml:"ttl"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env-default:"localhost:8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"4s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"4s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

const (
	minJWTSecretBytes = 32
	maxJWTTTL         = 24 * time.Hour
)

func MustLoad() *Config {
	configPath := os.Getenv("CS_CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CS_CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Cannot load config: %s", err)
	}
	cfg.StoragePath = os.Getenv("STORAGE_PATH")
	if cfg.StoragePath == "" {
		log.Fatal("STORAGE_PATH is not set")
	}
	if err := validate(&cfg); err != nil {
		log.Fatalf("invalid configuration: %s", err)
	}

	return &cfg
}

func validate(cfg *Config) error {
	secret := cfg.JWT.Secret
	trimmedSecret := strings.TrimSpace(secret)

	if trimmedSecret == "" {
		return fmt.Errorf("JWT secret is empty")
	}
	if secret != trimmedSecret {
		return fmt.Errorf("JWT secret must not have leading or trailing whitespace")
	}
	if strings.EqualFold(secret, "change-me") {
		return fmt.Errorf("default JWT secret must be replaced")
	}
	if len([]byte(secret)) < minJWTSecretBytes {
		return fmt.Errorf("JWT secret must contain at least %d bytes", minJWTSecretBytes)
	}
	if cfg.JWT.TTL <= 0 {
		return fmt.Errorf("JWT TTL must be positive")
	}
	if cfg.JWT.TTL > maxJWTTTL {
		return fmt.Errorf("JWT TTL must not exceed %s", maxJWTTTL)
	}

	return nil
}
