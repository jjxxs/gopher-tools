package config

import "os"

// GetEnvironmentOrDefault returns the environment-variable named by key
// or default if a variable with the key does not exist.
func GetEnvironmentOrDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return def
}
