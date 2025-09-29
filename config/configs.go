package config

import (
	"ermes-central-repo/logger"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
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
	ClientSecret            string        `mapstructure:"client_secret"`
	DevSecret               string        `mapstructure:"dev_secret"`
	FunctionRegistryAddress string        `mapstructure:"function_registry_address"`
	FunctionRegistryPort    string        `mapstructure:"function_registry_port"`
	FunctionCacheSize       int           `mapstructure:"function_cache_size"`
	FunctionCacheBinDir     string        `mapstructure:"function_cache_bin_dir"`
	FunctionCacheTtl        time.Duration `mapstructure:"function_cache_ttl"`
}

func LoadConfig(log *logger.Logger, cfgPath string) (*Config, error) {
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("Error getting user home directory: %v", err)
		return nil, fmt.Errorf("error getting user home directory: %v", err)
	}
	if cfgPath != "" {
		v.AddConfigPath(cfgPath)
	}
	v.AddConfigPath(filepath.Join(homeDir, ".ermes-cr"))
	v.AddConfigPath("/etc/ermes-cr/")
	v.AddConfigPath(".")

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
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("config.yaml file not found in any defined paths. Fallback to ENV variables and default values: %v", err)
			return nil, fmt.Errorf("config.yaml file not found in any defined paths: %v", err)
		} else {
			log.Errorf("Could not load config file: %v", err)
			return nil, fmt.Errorf("could not load config file: %v", err)
		}
	}
	log.Infof("Using config file: %s", v.ConfigFileUsed())

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
	if cfg.ClientSecret == "" {
		missing = append(missing, "client_secret")
	}
	if cfg.DevSecret == "" {
		missing = append(missing, "dev_secret")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required config values: %s", missing)
	}
	return &cfg, nil
}
