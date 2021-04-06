package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/poopmail/canalization/internal/shared"
)

// accountService represents the postgres account service implementation
type accountService struct {
	pool *pgxpool.Pool
}

// Count counts the total amount of accounts stored inside the database
func (service *accountService) Count() (int, error) {
	query := "SELECT COUNT(*) FROM accounts"

	row := service.pool.QueryRow(context.Background(), query)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Accounts retrieves the desired amount of accounts out of the database
func (service *accountService) Accounts(skip, limit int) ([]*shared.Account, error) {
	query := fmt.Sprintf("SELECT * FROM accounts ORDER BY created LIMIT %d OFFSET %d", limit, skip)

	rows, err := service.pool.Query(context.Background(), query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*shared.Account{}, nil
		}
		return nil, err
	}

	var accounts []*shared.Account
	for rows.Next() {
		account, err := rowToAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Account retrieves a specific account out of the database
func (service *accountService) Account(id snowflake.ID) (*shared.Account, error) {
	query := "SELECT * FROM accounts WHERE id = $1"

	account, err := rowToAccount(service.pool.QueryRow(context.Background(), query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return account, nil
}

// AccountByUsername retrieves a specific account with a specific username out of the database
func (service *accountService) AccountByUsername(username string) (*shared.Account, error) {
	query := "SELECT * FROM accounts WHERE username = $1"

	account, err := rowToAccount(service.pool.QueryRow(context.Background(), query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return account, nil
}

// CreateOrReplace creates or replaces an account inside the database
func (service *accountService) CreateOrReplace(account *shared.Account) error {
	query := `
		INSERT INTO accounts (id, username, password, admin, created)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
			SET username = excluded.username,
				password = excluded.password,
				admin = excluded.admin,
				created = excluded.created
	`

	_, err := service.pool.Exec(context.Background(), query, account.ID, account.Username, account.Password, account.Admin, account.Created)
	return err
}

// Delete deletes a specific account out of the database
func (service *accountService) Delete(id snowflake.ID) error {
	query := "DELETE FROM accounts WHERE id = $1"

	_, err := service.pool.Exec(context.Background(), query, id)
	return err
}

func rowToAccount(row pgx.Row) (*shared.Account, error) {
	account := new(shared.Account)

	if err := row.Scan(&account.ID, &account.Username, &account.Password, &account.Admin, &account.Created); err != nil {
		return nil, err
	}

	return account, nil
}
