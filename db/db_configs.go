package db

type RedisConfig struct {
	Addr     string
	Port     string
	Password string
	DB       int
	Protocol int
}

func GetRedisConfig() (RedisConfig, error) {
	return RedisConfig{
		Addr:     "redis",
		Port:     "6379",
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       0,
	}, nil
}
