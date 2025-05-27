package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"log"
	"sync"
)

var (
	cfg  *config
	once sync.Once
)

type config struct {
	JWTSecret     string `envconfig:"JWT_SECRET" required:"true"`
	RedisAddr     string `envconfig:"REDIS_ADDR" default:"redis"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`
}

func loadConfig() {
	cfg = &config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println(cfg)
}

func JWTSecret() string {
	once.Do(loadConfig)
	return cfg.JWTSecret
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
