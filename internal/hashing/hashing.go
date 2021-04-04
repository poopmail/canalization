package hashing

import "github.com/alexedwards/argon2id"

// Hash hashes the given password
func Hash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

// Check checks the given password by comparing it to the given hash
func Check(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
