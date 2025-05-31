package config

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
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
	fmt.Println(cfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println(cfg)
}

func JWTSecret() *rsa.PrivateKey {
	once.Do(loadConfig)
	signBytes, err := os.ReadFile(cfg.JWTSecret)
	if err != nil {
		log.Fatal("Error reading JWT secret file")
	}
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal("Error parsing JWT secret file")
	}

	return signKey
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
