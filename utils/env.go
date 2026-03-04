package utils

import (
	"os"
	"strings"
)

// LoadEnvFile reads a .env file and sets environment variables.
// It skips empty lines and comment lines starting with '#'.
// Existing environment variables are not overwritten.
func LoadEnvFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// .env file is optional, silently ignore if not found
		return
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already defined in the environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// GetEnv returns the value of the environment variable named by key.
// If the variable is not set or is empty, it returns the fallback value.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
