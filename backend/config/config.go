package config

import (
	"os"
	"strings"

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

// GetSchemaSourceMode retrieves the schema source mode from environment variable
// Context can be "schema" or "validation" (case-insensitive)
// - For "schema": reads from SCHEMA_SOURCE_MODE env var
// - For "validation": reads from VALIDATION_SOURCE_MODE env var
// Supports values: "local-then-bsr", "bsr-only", "local-only" (case-insensitive)
// Defaults to LocalThenBSR if not set or invalid
func GetSchemaSourceMode(context string) SchemaSourceMode {
	context = strings.ToLower(strings.TrimSpace(context))

	var envVar string
	var defaultMode string = "local-then-bsr"

	switch context {
	case "schema":
		envVar = "SCHEMA_SOURCE_MODE"
	case "validation":
		envVar = "VALIDATION_SOURCE_MODE"
	default:
		// Default to schema if invalid context
		envVar = "SCHEMA_SOURCE_MODE"
	}

	modeStr := GetEnv(envVar, defaultMode)
	modeStr = strings.ToLower(strings.TrimSpace(modeStr))

	switch modeStr {
	case "bsr-only":
		return BSROnly
	case "local-only":
		return LocalOnly
	case "local-then-bsr":
		return LocalThenBSR
	default:
		// Default to LocalThenBSR for invalid values
		return LocalThenBSR
	}
}

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
