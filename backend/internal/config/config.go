package config

import "os"

type Config struct {
	HTTPAddress       string
	DatabaseURL       string
	MarketDataBaseURL string
	ClerkIssuerURL    string
	ClerkJWKSURL      string
	AuthMockMode      bool
}

func Load() Config {
	return Config{
		HTTPAddress:       getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://tradelab:tradelab@localhost:5432/tradelab?sslmode=disable"),
		MarketDataBaseURL: getEnv("MARKET_DATA_BASE_URL", "https://api.binance.com"),
		ClerkIssuerURL:    getEnvAny([]string{"TRADESLAB_CLERK_ISSUER_URL", "CLERK_ISSUER_URL"}, ""),
		ClerkJWKSURL:      getEnvAny([]string{"TRADESLAB_CLERK_JWKS_URL", "CLERK_JWKS_URL"}, ""),
		AuthMockMode:      getEnv("TRADESLAB_AUTH_MOCK_MODE", "") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvAny(keys []string, fallback string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}

	return fallback
}
