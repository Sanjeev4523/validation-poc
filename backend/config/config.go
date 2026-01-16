package config

import (
	"os"

	"github.com/joho/godotenv"
)

// SchemaSourceMode defines the strategy for retrieving schemas
type SchemaSourceMode int

const (
	// LocalThenBSR checks local files first, then falls back to BSR if not found
	LocalThenBSR SchemaSourceMode = iota
	// BSROnly skips local file check and fetches directly from BSR
	BSROnly
	// LocalOnly only uses local files, never fetches from BSR
	LocalOnly
)

// SCHEMA_SOURCE_MODE controls how schemas are retrieved
// Change this constant to switch between modes:
// - BSROnly: Fetch directly from BSR (skip local check)
// - LocalOnly: Only use local files (never fetch from BSR)
// - LocalThenBSR: Check local first, then BSR (default behavior)
const SCHEMA_SOURCE_MODE = BSROnly

// LoadEnv loads environment variables from .env file
// If the .env file doesn't exist, it silently falls back to system environment variables
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		// .env file doesn't exist, which is okay - we'll use system env vars
		return nil
	}
	return nil
}

// GetEnv retrieves an environment variable with a default value
// It first checks the environment (which may have been loaded from .env file)
// If not found, returns the default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
