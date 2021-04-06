package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/poopmail/canalization/internal/shared"
)

// refreshTokenService represents the postgres refresh token service implementation
type refreshTokenService struct {
	pool *pgxpool.Pool
}

// Count counts the total amount of refresh tokens of a specific account stored inside the database
func (service *refreshTokenService) Count(account snowflake.ID) (int, error) {
	query := "SELECT COUNT(*) FROM refresh_tokens WHERE account = $1"

	row := service.pool.QueryRow(context.Background(), query, account)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// RefreshTokens retrieves the desired amount of refresh tokens of a specific amount out of the database
func (service *refreshTokenService) RefreshTokens(account snowflake.ID, skip, limit int) ([]*shared.RefreshToken, error) {
	query := fmt.Sprintf("SELECT * FROM refresh_tokens WHERE account = $1 ORDER BY created LIMIT %d OFFSET %d", limit, skip)

	rows, err := service.pool.Query(context.Background(), query, account)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*shared.RefreshToken{}, nil
		}
		return nil, err
	}

	var refreshTokens []*shared.RefreshToken
	for rows.Next() {
		refreshToken, err := rowToRefreshToken(rows)
		if err != nil {
			return nil, err
		}
		refreshTokens = append(refreshTokens, refreshToken)
	}

	return refreshTokens, nil
}

// RefreshToken retrieves a specific refresh token from a specific account out of the database
func (service *refreshTokenService) RefreshToken(account, id snowflake.ID) (*shared.RefreshToken, error) {
	query := "SELECT * FROM refresh_tokens WHERE id = $1 AND account = $2"

	refreshToken, err := rowToRefreshToken(service.pool.QueryRow(context.Background(), query, id, account))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return refreshToken, nil
}

// CreateOrReplace creates or replaces a refresh token inside the database
func (service *refreshTokenService) CreateOrReplace(token *shared.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, account, token, description, created)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id, account) DO UPDATE
			SET token = excluded.token,
				description = excluded.description,
				created = excluded.created
	`

	_, err := service.pool.Exec(context.Background(), query, token.ID, token.Account, token.Token, token.Description, token.Created)
	return err
}

// Delete deletes a specific refresh token from a specific account out of the database
func (service *refreshTokenService) Delete(account, id snowflake.ID) error {
	query := "DELETE FROM refresh_tokens WHERE id = $1 AND account = $2"

	_, err := service.pool.Exec(context.Background(), query, id, account)
	return err
}

// DeleteAll deletes all refresh tokens from a specific account out of the database
func (service *refreshTokenService) DeleteAll(account snowflake.ID) error {
	query := "DELETE FROM refresh_tokens WHERE account = $1"

	_, err := service.pool.Exec(context.Background(), query, account)
	return err
}

// DeleteExpired deletes all expired refresh tokens
func (service *refreshTokenService) DeleteExpired(valid time.Duration) (int64, error) {
	query := "DELETE FROM refresh_tokens WHERE created < $1"

	tag, err := service.pool.Exec(context.Background(), query, time.Now().Add(-valid).Unix())
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func rowToRefreshToken(row pgx.Row) (*shared.RefreshToken, error) {
	refreshToken := new(shared.RefreshToken)

	if err := row.Scan(&refreshToken.ID, &refreshToken.Account, &refreshToken.Token, &refreshToken.Description, &refreshToken.Created); err != nil {
		return nil, err
	}

	return refreshToken, nil
}
