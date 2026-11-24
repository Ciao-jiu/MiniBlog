package config

import "os"

// Env returns an environment variable or a local-development fallback.
func Env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
