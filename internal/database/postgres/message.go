package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/poopmail/canalization/internal/shared"
)

// messageService represents the postgres message service implementation
type messageService struct {
	pool *pgxpool.Pool
}

// Count counts the total amount of messages in a specific mailbox stored inside the database
func (service *messageService) Count(mailbox string) (int, error) {
	query := "SELECT COUNT(*) FROM messages WHERE mailbox = $1"

	row := service.pool.QueryRow(context.Background(), query, strings.ToLower(mailbox))

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Messages retrieves the desired amount of messages in a specific mailbox out of the database
func (service *messageService) Messages(mailbox string, skip, limit int) ([]*shared.Message, error) {
	query := fmt.Sprintf("SELECT * FROM messages WHERE mailbox = $1 ORDER BY created LIMIT %d OFFSET %d", limit, skip)

	rows, err := service.pool.Query(context.Background(), query, strings.ToLower(mailbox))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*shared.Message{}, nil
		}
		return nil, err
	}

	var messages []*shared.Message
	for rows.Next() {
		message, err := rowToMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// Message retrieves a specific message with a specific ID out of the database
func (service *messageService) Message(id snowflake.ID) (*shared.Message, error) {
	query := "SELECT * FROM messages WHERE id = $1"

	message, err := rowToMessage(service.pool.QueryRow(context.Background(), query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return message, nil
}

// CreateOrReplace creates or replaces a message inside the database
func (service *messageService) CreateOrReplace(message *shared.Message) error {
	query := `
		INSERT INTO messages (id, mailbox, "from", subject, content_plain, content_html, created)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
			SET mailbox = excluded.mailbox,
				"from" = excluded.from,
				subject = excluded.subject,
				content_plain = excluded.content_plain,
				content_html = excluded.content_html,
				created = excluded.created
	`

	_, err := service.pool.Exec(context.Background(), query, message.ID, strings.ToLower(message.Mailbox), message.From, message.Subject, message.Content.Plain, message.Content.HTML, message.Created)
	return err
}

// Delete deletes a specific message with a specific ID out of the database
func (service *messageService) Delete(id snowflake.ID) error {
	query := "DELETE FROM messages WHERE id = $1"

	_, err := service.pool.Exec(context.Background(), query, id)
	return err
}

// DeleteInMailbox deletes all messages in a specific mailbox out of the database
func (service *messageService) DeleteInMailbox(mailbox string) error {
	query := "DELETE FROM messages WHERE mailbox = $1"

	_, err := service.pool.Exec(context.Background(), query, strings.ToLower(mailbox))
	return err
}

func rowToMessage(row pgx.Row) (*shared.Message, error) {
	message := new(shared.Message)
	message.Content = new(shared.MessageContent)

	if err := row.Scan(&message.ID, &message.Mailbox, &message.From, &message.Subject, &message.Content.Plain, &message.Content.HTML, &message.Created); err != nil {
		return nil, err
	}

	return message, nil
}
