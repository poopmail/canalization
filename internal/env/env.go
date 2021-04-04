package env

import (
	"os"
	"strconv"
)

// MustString returns the string set under the given environment variable key or the fallback if it is not set
func MustString(key, fallback string) string {
	value, set := os.LookupEnv(key)
	if !set {
		return fallback
	}
	return value
}

// MustInt returns the integer set under the given environment variable key or the fallback if it is not set or cannot be parsed
func MustInt(key string, fallback int) int {
	value, set := os.LookupEnv(key)
	if !set {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// MustBool returns the boolean set under the given environment variable key or the fallback if it is not set or cannot be parsed
func MustBool(key string, fallback bool) bool {
	value, set := os.LookupEnv(key)
	if !set {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
