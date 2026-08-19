package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/numbergroup/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
	// StoragePath string  `yaml:"storage_path" env-required:"true"`
	StoragePath string
	StorageKey  string
	HTTPServer  `yaml:"http_server"`
	JWT         JWT `yaml:"jwt"`
}

func (c Config) String() string {
	return fmt.Sprintf(
		"{Env:%s StoragePath:%s SQLiteKey:*** HTTPServer:%+v}",
		c.Env,
		c.StoragePath,
		c.HTTPServer,
	)
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

	cfg.StorageKey = os.Getenv("SQLITE_KEY")
	if cfg.StorageKey == "" {
		log.Fatal("SQLITE_KEY is not set")
	}
	return &cfg
}
