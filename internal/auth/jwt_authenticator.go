package auth

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/dgrijalva/jwt-go"
	"github.com/poopmail/canalization/internal/shared"
)

// JWTAuthenticator represents an authenticator implementation for JWTs
type JWTAuthenticator struct {
	signingKey []byte
	lifetime   time.Duration
	accounts   shared.AccountService
}

// NewJWTAuthenticator creates a new JWT authenticator
func NewJWTAuthenticator(signingKey string, lifetime time.Duration, accounts shared.AccountService) *JWTAuthenticator {
	return &JWTAuthenticator{
		signingKey: []byte(signingKey),
		lifetime:   lifetime,
		accounts:   accounts,
	}
}

// GenerateToken generates a new signed JWT for a given account
func (authenticator *JWTAuthenticator) GenerateToken(account *shared.Account) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: now.Add(authenticator.lifetime).Unix(),
		IssuedAt:  now.Unix(),
		Subject:   account.ID.String(),
	}).SignedString(authenticator.signingKey)
}

// ProcessToken validates a given JWT and retrieves the corresponding account for it
func (authenticator *JWTAuthenticator) ProcessToken(token string) (bool, *shared.Account, error) {
	// Try to parse the JWT
	claims := new(jwt.StandardClaims)
	jwtToken, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (interface{}, error) {
		return authenticator.signingKey, nil
	})
	if err != nil {
		return false, nil, err
	}

	// Check if the JWT itself is valid
	if !jwtToken.Valid {
		return false, nil, nil
	}

	// Try to parse the given subject to a snowflake ID
	id, err := snowflake.ParseString(claims.Subject)
	if err != nil {
		return false, nil, nil
	}

	// Try to retrieve the account stored under the parsed snowflake ID
	account, err := authenticator.accounts.Account(id)
	if err != nil {
		return false, nil, err
	}
	if account == nil {
		return false, nil, nil
	}

	// Fail if the account executed a token reset after the JWT was issued
	if claims.IssuedAt < account.TokenReset {
		return false, nil, nil
	}

	return true, account, nil
}
