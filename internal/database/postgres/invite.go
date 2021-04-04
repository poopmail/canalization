package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/poopmail/canalization/internal/shared"
)

// inviteService represents the postgres invite service implementation
type inviteService struct {
	pool *pgxpool.Pool
}

// Count counts the total amount of invites stored inside the database
func (service *inviteService) Count() (int, error) {
	query := "SELECT COUNT(*) FROM invites"

	row := service.pool.QueryRow(context.Background(), query)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Invites retrieves the desired amount of invites out of the namespace
func (service *inviteService) Invites(skip, limit int) ([]*shared.Invite, error) {
	query := fmt.Sprintf("SELECT * FROM invites ORDER BY created LIMIT %d OFFSET %d", limit, skip)

	rows, err := service.pool.Query(context.Background(), query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*shared.Invite{}, nil
		}
		return nil, err
	}

	var invites []*shared.Invite
	for rows.Next() {
		invite, err := rowToInvite(rows)
		if err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}

	return invites, nil
}

// Invite retrieves a specific invite with a specific code out of the database
func (service *inviteService) Invite(code string) (*shared.Invite, error) {
	query := "SELECT * FROM invites WHERE code = $1"

	invite, err := rowToInvite(service.pool.QueryRow(context.Background(), query, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return invite, nil
}

// CreateOrReplace creates or replaces an invite inside the database
func (service *inviteService) CreateOrReplace(invite *shared.Invite) error {
	query := `
		INSERT INTO invites (code, created)
		VALUES ($1, $2)
		ON CONFLICT (code) DO UPDATE
			SET created = excluded.created
	`

	_, err := service.pool.Exec(context.Background(), query, invite.Code, invite.Created)
	return err
}

// Delete deletes a specific invite with a specific code out of the database
func (service *inviteService) Delete(code string) error {
	query := "DELETE FROM invites WHERE code = $1"

	_, err := service.pool.Exec(context.Background(), query, code)
	return err
}

func rowToInvite(row pgx.Row) (*shared.Invite, error) {
	invite := new(shared.Invite)

	if err := row.Scan(&invite.Code, &invite.Created); err != nil {
		return nil, err
	}

	return invite, nil
}
