package shared

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// RefreshToken represents an accounts refresh token
type RefreshToken struct {
	ID          snowflake.ID `json:"id"`
	Account     snowflake.ID `json:"account"`
	Token       string       `json:"token,omitempty"`
	Description string       `json:"description"`
	Created     int64        `json:"created"`
}

// RefreshTokenService represents a service which keeps track of account refresh tokens
type RefreshTokenService interface {
	Count(account snowflake.ID) (int, error)
	RefreshTokens(account snowflake.ID, skip, limit int) ([]*RefreshToken, error)
	RefreshToken(account snowflake.ID, id snowflake.ID) (*RefreshToken, error)
	CreateOrReplace(token *RefreshToken) error
	Delete(account snowflake.ID, id snowflake.ID) error
	DeleteAll(account snowflake.ID) error
	DeleteExpired(valid time.Duration) (int64, error)
}
