package config

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	cfg  *config
	once sync.Once
)

type config struct {
	JWTPemPath              string `envconfig:"JWT_PEM_PATH" required:"true"`
	JWTExpirationTime       int    `envconfig:"JWT_EXPIRATION_TIME" default:"3600"`
	RedisAddr               string `envconfig:"REDIS_ADDR" default:"redis"`
	RedisPort               string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword           string `envconfig:"REDIS_PASSWORD"`
	RedisDB                 int    `envconfig:"REDIS_DB" default:"0"`
	HTTPPort                string `envconfig:"HTTP_PORT" default:"8080"`
	WASMPoolSize            int    `envconfig:"WASM_POOL_SIZE" default:"10"`
	ErmesSecret             string `envconfig:"ERMES_SECRET" required:"true"`
	MiddlewareSecret        string `envconfig:"MIDDLEWARE_SECRET" required:"true"`
	DevSecret               string `envconfig:"DEV_SECRET" required:"true"`
	FunctionRegistryAddress string `envconfig:"FUNCTION_REGISTRY_ADDRESS" default:"localhost"`
	FunctionRegistryPort    string `envconfig:"FUNCTION_REGISTRY_PORT" default:"8080"`
	FunctionCacheSize       int    `envconfig:"FUNCTION_CACHE_SIZE" default:"100"`
	FunctionCacheBinDir     string `envconfig:"FUNCTION_CACHE_BIN_DIR" default:"./function_cache"`
	FunctionCacheTtl        int    `envconfig:"FUNCTION_CACHE_TTL" default:"0"`
}

func loadConfig() {
	cfg = &config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println(cfg)
}

func JWTPemPath() string {
	once.Do(loadConfig)
	return cfg.JWTPemPath
}

func JWTExpirationTime() time.Duration {
	once.Do(loadConfig)
	return time.Duration(cfg.JWTExpirationTime)
}

func RedisAddr() string {
	once.Do(loadConfig)
	return cfg.RedisAddr
}

func RedisPort() string {
	once.Do(loadConfig)
	return cfg.RedisPort
}

func RedisPassword() string {
	once.Do(loadConfig)
	return cfg.RedisPassword
}

func RedisDB() int {
	once.Do(loadConfig)
	return cfg.RedisDB
}

func HTTPPort() string {
	once.Do(loadConfig)
	return cfg.HTTPPort
}

func WASMPoolSize() int {
	once.Do(loadConfig)
	return cfg.WASMPoolSize
}

func ErmesSecret() string {
	once.Do(loadConfig)
	return cfg.ErmesSecret
}

func MiddlewareSecret() string {
	once.Do(loadConfig)
	return cfg.MiddlewareSecret
}

func DevSecret() string {
	once.Do(loadConfig)
	return cfg.DevSecret
}

func FunctionRegistryAddress() string {
	once.Do(loadConfig)
	return cfg.FunctionRegistryAddress
}

func FunctionRegistryPort() string {
	once.Do(loadConfig)
	return cfg.FunctionRegistryPort
}

func FunctionCacheSize() int {
	once.Do(loadConfig)
	return cfg.FunctionCacheSize
}

func FunctionCacheBinDir() string {
	once.Do(loadConfig)
	return cfg.FunctionCacheBinDir
}

func FunctionCacheTtl() time.Duration {
	once.Do(loadConfig)
	return time.Duration(cfg.FunctionCacheTtl)
}
