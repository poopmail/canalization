package random

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	allowedCharacters      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+*#-_.,"
	allowedCharactersRunes = []rune(allowedCharacters)
)

// RandomString generates a random string with a specified length
func RandomString(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = allowedCharactersRunes[rand.Intn(len(allowedCharactersRunes))]
	}
	return string(runes)
}
