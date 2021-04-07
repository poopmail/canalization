package validation

import (
	"strings"
)

var allowedMailboxKeyCharacters = "üäöÜÄÖabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// ValidateMailboxKey validates a mailbox key
func ValidateMailboxKey(key string) bool {
	for _, char := range key {
		if !strings.ContainsRune(allowedMailboxKeyCharacters, char) {
			return false
		}
	}

	return true
}
