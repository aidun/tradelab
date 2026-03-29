package config

import "os"

type Config struct {
	HTTPAddress       string
	DatabaseURL       string
	MarketDataBaseURL string
}

func Load() Config {
	return Config{
		HTTPAddress:       getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://tradelab:tradelab@localhost:5432/tradelab?sslmode=disable"),
		MarketDataBaseURL: getEnv("MARKET_DATA_BASE_URL", "https://api.binance.com"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
