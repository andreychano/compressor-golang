package config

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/shanth1/gotools/conf"
	"github.com/shanth1/gotools/consts"
	"github.com/shanth1/gotools/env"
	"github.com/shanth1/gotools/flags"
)

type bootstrapConfig struct {
	AppEnv     string `flag:"env" usage:"Environment: local, dev, stage, prod"`
	ConfigPath string `flag:"config" usage:"Path to YAML config file"`
	EnvPath    string `flag:"env-path" usage:"Path to .env file"`
}

func Load(ctx context.Context) (*Config, error) {
	boot := &bootstrapConfig{}
	if err := flags.RegisterFromStruct(boot); err != nil {
		return nil, fmt.Errorf("register flags: %w", err)
	}
	flag.Parse()

	// Приоритет: flag > APP_ENV > "local"
	appEnv := boot.AppEnv
	if appEnv == "" {
		if envVar, exists := os.LookupEnv("APP_ENV"); exists {
			appEnv = envVar
		} else {
			appEnv = "local"
		}
	}

	// Универсальный root path
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("cannot get runtime caller")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	// Config path по умолчанию
	if boot.ConfigPath == "" {
		boot.ConfigPath = filepath.Join(root, "internal/config", fmt.Sprintf("config.%s.yaml", appEnv))
	}

	cfg := &Config{}

	// Проверка существования
	if _, err := os.Stat(boot.ConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config not found: %s\nls -la %s/internal/config/", boot.ConfigPath, root)
	}

	// Загрузка YAML
	if err := conf.Load(boot.ConfigPath, cfg); err != nil {
		return nil, fmt.Errorf("load YAML %s: %w", boot.ConfigPath, err)
	}

	// .env override (тихо)
	if boot.EnvPath != "" {
		if err := env.LoadIntoStruct(boot.EnvPath, cfg); err != nil {
			log.Printf("Warning: .env %s failed: %v", boot.EnvPath, err)
		}
	}

	cfg.Env = consts.Env(appEnv)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	log.Printf("Config loaded: env=%s, HTTP=%s, LogLevel=%s", appEnv, cfg.HTTP.Address, cfg.Log.Level)
	return cfg, nil
}

func MustLoad(ctx context.Context) *Config {
	cfg, err := Load(ctx)
	if err != nil {
		log.Fatalf("Config failed: %v", err)
	}
	return cfg
}
