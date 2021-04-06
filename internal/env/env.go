package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// MustString returns the string set under the given environment variable key or the fallback if it is not set
func MustString(key, fallback string) string {
	value, set := os.LookupEnv(key)
	if !set {
		return fallback
	}
	return value
}

// MustStringSlice returns the split string using the given separator set under the given environment variable key or the fallback slice if it is not set
func MustStringSlice(key, separator string, fallback []string) []string {
	value, set := os.LookupEnv(key)
	if !set {
		return fallback
	}
	return strings.Split(value, separator)
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

// MustDuration returns the duration set under the given environment variable key or the fallback if it is not set or cannot be parsed
func MustDuration(key string, negativeAllowed bool, fallback time.Duration) time.Duration {
	value, set := os.LookupEnv(key)
	if !set {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	if duration < 0 && !negativeAllowed {
		return fallback
	}

	return duration
}
