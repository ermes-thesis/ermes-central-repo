package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"log"
	"sync"
	"time"
)

var (
	cfg  *config
	once sync.Once
)

type config struct {
	JWTPemPath        string `envconfig:"JWT_PEM_PATH" required:"true"`
	JWTExpirationTime int    `envconfig:"JWT_EXPIRATION_TIME" default:"3600"`
	RedisAddr         string `envconfig:"REDIS_ADDR" default:"redis"`
	RedisPort         string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword     string `envconfig:"REDIS_PASSWORD"`
	RedisDB           int    `envconfig:"REDIS_DB" default:"0"`
	HTTPPort          string `envconfig:"HTTP_PORT" default:"8080"`
	WASMPoolSize      int    `envconfig:"WASM_POOL_SIZE" default:"10"`
	ErmesSecret       string `envconfig:"ERMES_SECRET" required:"true"`
	MiddlewareSecret  string `envconfig:"MIDDLEWARE_SECRET" required:"true"`
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
