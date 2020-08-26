package config

import "os"

// Returns the environment-variable named by key or default if
// it does not exist.
func GetEnvironmentOrDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return def
}
