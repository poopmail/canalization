package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/poopmail/canalization/internal/shared"
)

// mailboxService represents the postgres mailbox service implementation
type mailboxService struct {
	pool *pgxpool.Pool
}

// Count counts the total amount of mailboxes stored inside the database
func (service *mailboxService) Count() (int, error) {
	query := "SELECT COUNT(*) FROM mailboxes"

	row := service.pool.QueryRow(context.Background(), query)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Mailboxes retrieves the desired amount of mailboxes out of the database
func (service *mailboxService) Mailboxes(skip, limit int) ([]*shared.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM mailboxes ORDER BY created LIMIT %d OFFSET %d", limit, skip)

	rows, err := service.pool.Query(context.Background(), query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*shared.Mailbox{}, nil
		}
		return nil, err
	}

	var mailboxes []*shared.Mailbox
	for rows.Next() {
		mailbox, err := rowToMailbox(rows)
		if err != nil {
			return nil, err
		}
		mailboxes = append(mailboxes, mailbox)
	}

	return mailboxes, nil
}

// CountInAccount counts the total amount of mailboxes in a specific account stored inside the database
func (service *mailboxService) CountInAccount(account string) (int, error) {
	query := "SELECT COUNT(*) FROM mailboxes WHERE account = $1"

	row := service.pool.QueryRow(context.Background(), query, account)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// MailboxesInAccount retrieves the desired amount of mailboxes in a specific account out of the database
func (service *mailboxService) MailboxesInAccount(account string, skip, limit int) ([]*shared.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM mailboxes WHERE account = $1 ORDER BY created LIMIT %d OFFSET %d", limit, skip)

	rows, err := service.pool.Query(context.Background(), query, account)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*shared.Mailbox{}, nil
		}
		return nil, err
	}

	var mailboxes []*shared.Mailbox
	for rows.Next() {
		mailbox, err := rowToMailbox(rows)
		if err != nil {
			return nil, err
		}
		mailboxes = append(mailboxes, mailbox)
	}

	return mailboxes, nil
}

// Mailbox retrieves a specific mailbox with a specific address out of the database
func (service *mailboxService) Mailbox(address string) (*shared.Mailbox, error) {
	query := "SELECT * FROM mailboxes WHERE address = $1"

	mailbox, err := rowToMailbox(service.pool.QueryRow(context.Background(), query, address))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mailbox, nil
}

// CreateOrReplace creates or replaces a mailbox inside the database
func (service *mailboxService) CreateOrReplace(mailbox *shared.Mailbox) error {
	query := `
		INSERT INTO mailboxes (address, account, created)
		VALUES ($1, $2, $3)
		ON CONFLICT (address) DO UPDATE
			SET account = excluded.account,
				created = excluded.created
	`

	_, err := service.pool.Exec(context.Background(), query, mailbox.Address, mailbox.Account, mailbox.Created)
	return err
}

// Delete deletes a specific mailbox with a specific address out of the database
func (service *mailboxService) Delete(address string) error {
	query := "DELETE FROM mailboxes WHERE address = $1"

	_, err := service.pool.Exec(context.Background(), query, address)
	return err
}

// DeleteInAccount deletes all mailboxes in a specific account out of the database
func (service *mailboxService) DeleteInAccount(account string) error {
	query := "DELETE FROM mailboxes WHERE account = $1"

	_, err := service.pool.Exec(context.Background(), query, account)
	return err
}

func rowToMailbox(row pgx.Row) (*shared.Mailbox, error) {
	mailbox := new(shared.Mailbox)

	if err := row.Scan(&mailbox.Address, &mailbox.Account, &mailbox.Created); err != nil {
		return nil, err
	}

	return mailbox, nil
}
