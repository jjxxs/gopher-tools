package config

import (
	"fmt"
	"os"
)

// GetEnvironmentOrDefault returns the environment-variable named by key
// or default if a variable with the key does not exist.
func GetEnvironmentOrDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

// GetEnvironmentOrPanic returns the environment-variable named by key
// or panics if a variable with the key does not exist.
func GetEnvironmentOrPanic(key string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	panic(fmt.Errorf("environment variable for %s not set", key))
}
