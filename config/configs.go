package config

import (
	"ermes-central-repo/logger"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Config struct {
	JWTPemPath              string        `mapstructure:"jwt_pem_path"`
	JWTExpirationTime       time.Duration `mapstructure:"jwt_expiration_time"`
	RedisAddr               string        `mapstructure:"redis_addr"`
	RedisPort               string        `mapstructure:"redis_port"`
	RedisPassword           string        `mapstructure:"redis_password"`
	RedisDB                 int           `mapstructure:"redis_db"`
	HTTPPort                string        `mapstructure:"http_port"`
	WASMPoolSize            int           `mapstructure:"wasm_pool_size"`
	ErmesSecret             string        `mapstructure:"ermes_secret"`
	MiddlewareSecret        string        `mapstructure:"middleware_secret"`
	DevSecret               string        `mapstructure:"dev_secret"`
	FunctionRegistryAddress string        `mapstructure:"function_registry_address"`
	FunctionRegistryPort    string        `mapstructure:"function_registry_port"`
	FunctionCacheSize       int           `mapstructure:"function_cache_size"`
	FunctionCacheBinDir     string        `mapstructure:"function_cache_bin_dir"`
	FunctionCacheTtl        time.Duration `mapstructure:"function_cache_ttl"`
}

func LoadConfig(cfgPath string, log *logger.Logger) (*Config, error) {
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	if cfgPath == "" {
		cfgPath = "./config.yml"
	}
	log.Infof("Using config file: %s", cfgPath)

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", cfgPath)
	}

	v := viper.New()

	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")

	v.SetDefault("jwt_expiration_time", "3600s")
	v.SetDefault("redis_addr", "redis")
	v.SetDefault("redis_port", "6379")
	v.SetDefault("redis_db", "0")
	v.SetDefault("http_port", "8080")
	v.SetDefault("wasm_pool_size", 10)
	v.SetDefault("function_registry_address", "localhost")
	v.SetDefault("function_registry_port", "8080")
	v.SetDefault("function_cache_size", 100)
	v.SetDefault("function_cache_bin_dir", "./function_cache")
	v.SetDefault("function_cache_ttl", 0)

	v.AutomaticEnv()

	v.SetEnvPrefix("")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %s", err)
	}

	var missing []string
	if cfg.JWTPemPath == "" {
		missing = append(missing, "jwt_pem_path")
	}
	if cfg.ErmesSecret == "" {
		missing = append(missing, "ermes_secret")
	}
	if cfg.MiddlewareSecret == "" {
		missing = append(missing, "middleware_secret")
	}
	if cfg.DevSecret == "" {
		missing = append(missing, "dev_secret")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required config values: %s", missing)
	}
	return &cfg, nil
}
