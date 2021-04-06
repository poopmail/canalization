package postgres

import (
	"context"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/johejo/golang-migrate-extra/source/iofs"
)

//go:embed migrations/*.sql
var migrations embed.FS

// postgresDriver represents the postgres database driver
type postgresDriver struct {
	dsn           string
	pool          *pgxpool.Pool
	Accounts      *accountService
	RefreshTokens *refreshTokenService
	Invites       *inviteService
	Mailboxes     *mailboxService
	Messages      *messageService
}

// NewDriver creates a new postgres database driver
func NewDriver(dsn string) (*postgresDriver, error) {
	// Open a postgres connection pool
	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	// Create and return the postgres driver
	return &postgresDriver{
		dsn:           dsn,
		pool:          pool,
		Accounts:      &accountService{pool: pool},
		RefreshTokens: &refreshTokenService{pool: pool},
		Invites:       &inviteService{pool: pool},
		Mailboxes:     &mailboxService{pool: pool},
		Messages:      &messageService{pool: pool},
	}, nil
}

// Migrate runs all migrations on the connected database
func (driver *postgresDriver) Migrate() error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", source, driver.dsn)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Close closes the postgres database driver
func (driver *postgresDriver) Close() {
	driver.pool.Close()
}
