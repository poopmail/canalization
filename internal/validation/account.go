package validation

import (
	"strings"
	"unicode/utf8"
)

var (
	minAccountNameLength         = 3
	maxAccountNameLength         = 32
	allowedAccountNameCharacters = "üäöÜÄÖabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

// ValidateAccountName validates an account name
func ValidateAccountName(name string) bool {
	if utf8.RuneCountInString(name) < minAccountNameLength {
		return false
	}

	if utf8.RuneCountInString(name) > maxAccountNameLength {
		return false
	}

	for _, char := range name {
		if !strings.ContainsRune(allowedAccountNameCharacters, char) {
			return false
		}
	}

	return true
}
