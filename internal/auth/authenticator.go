package auth

import "github.com/poopmail/canalization/internal/shared"

// Authenticator represents an authenticator which authenticates users
type Authenticator interface {
	GenerateToken(account *shared.Account) (string, error)
	ProcessToken(token string) (bool, *shared.Account, error)
}
