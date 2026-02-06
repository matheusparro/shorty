package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	
	JWTSecret string
	BaseURL   string

	RedisAddr     string
	RedisPassword string
	RedisDB       string

	KafkaBrokers []string
	KafkaEnabled bool
}

func Load() Config {
	return Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "shorty"),
		DBPassword: getEnv("DB_PASSWORD", "shorty"),
		DBName:     getEnv("DB_NAME", "shorty"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),
		
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnv("REDIS_DB", "0"),

		KafkaBrokers: getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaEnabled: getEnvBool("KAFKA_ENABLED", true),
	}
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}