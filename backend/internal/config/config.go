package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
	Kafka    KafkaConfig
	Log      LogConfig
	Debug    bool
}

type ServerConfig struct {
	Name string
	Host string
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
	ShardCount   int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret     string
	Issuer     string
	ExpireHour int
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type KafkaConfig struct {
	Brokers []string
}

type LogConfig struct {
	Level  string
	Format string
	Output string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Name: env("SERVER_NAME", "doutok"),
			Host: env("SERVER_HOST", "0.0.0.0"),
			Port: envInt("SERVER_PORT", 8080),
			Mode: env("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:         env("DB_HOST", "127.0.0.1"),
			Port:         envInt("DB_PORT", 3306),
			User:         env("DB_USER", "doutok"),
			Password:     env("DB_PASSWORD", "doutok123"),
			DBName:       env("DB_NAME", "doutok"),
			MaxIdleConns: envInt("DB_MAX_IDLE", 10),
			MaxOpenConns: envInt("DB_MAX_OPEN", 100),
			MaxLifetime:  time.Hour,
			ShardCount:   envInt("DB_SHARD_COUNT", 16),
		},
		Redis: RedisConfig{
			Host:     env("REDIS_HOST", "127.0.0.1"),
			Port:     envInt("REDIS_PORT", 6379),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     env("JWT_SECRET", "doutok-change-me-in-prod"),
			Issuer:     "doutok",
			ExpireHour: envInt("JWT_EXPIRE_HOUR", 72),
		},
		MinIO: MinIOConfig{
			Endpoint:  env("MINIO_ENDPOINT", "127.0.0.1:9000"),
			AccessKey: env("MINIO_ACCESS_KEY", "doutok"),
			SecretKey: env("MINIO_SECRET_KEY", "doutok123"),
			Bucket:    env("MINIO_BUCKET", "doutok"),
		},
		Kafka: KafkaConfig{
			Brokers: []string{env("KAFKA_BROKERS", "127.0.0.1:9092")},
		},
		Log: LogConfig{
			Level:  env("LOG_LEVEL", "debug"),
			Format: env("LOG_FORMAT", "console"),
			Output: env("LOG_OUTPUT", "stdout"),
		},
		Debug: env("DEBUG", "true") == "true",
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
