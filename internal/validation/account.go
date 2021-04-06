package validation

import (
	"strings"
	"unicode/utf8"
)

var (
	minLength         = 3
	maxLength         = 32
	allowedCharacters = "üäöÜÄÖabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

// ValidateAccountName validates an account name
func ValidateAccountName(name string) bool {
	if utf8.RuneCountInString(name) < minLength {
		return false
	}

	if utf8.RuneCountInString(name) > maxLength {
		return false
	}

	for _, char := range name {
		if !strings.ContainsRune(allowedCharacters, char) {
			return false
		}
	}

	return true
}
