package db

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBname   string
}

type RedisConfig struct {
	Addr     string
	Port     string
	Password string
	DB       int
	Protocol int
}

func GetPostgresConfig() (PostgresConfig, error) {
	return PostgresConfig{
		Host:     "postgres",
		Port:     "5432",
		User:     "central_repo",
		Password: "password",
		DBname:   "central_repo",
	}, nil
}

func GetRedisConfig() (RedisConfig, error) {
	return RedisConfig{
		Addr:     "redis",
		Port:     "6379",
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       0,
	}, nil
}
