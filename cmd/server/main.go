package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/poopmail/canalization/internal/api"
	"github.com/poopmail/canalization/internal/config"
	"github.com/poopmail/canalization/internal/database/postgres"
	"github.com/poopmail/canalization/internal/hashing"
	"github.com/poopmail/canalization/internal/id"
	"github.com/poopmail/canalization/internal/karen"
	"github.com/poopmail/canalization/internal/shared"
	"github.com/poopmail/canalization/internal/static"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize the postgres database driver
	driver, err := postgres.NewDriver(config.Loaded.PostgresDSN)
	if err != nil {
		logrus.WithError(err).Fatal()
	}
	if err := driver.Migrate(); err != nil {
		logrus.WithError(err).Fatal()
	}

	pw, _ := hashing.Hash("a")
	driver.Accounts.CreateOrReplace(&shared.Account{
		ID:       id.Generate(),
		Username: "kse",
		Password: pw,
		Admin:    true,
		Created:  time.Now().Unix(),
	})

	// Start up the refresh token cleanup task
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go refreshTokenCleanup(ctx, driver.RefreshTokens, config.Loaded.RefreshTokenLifetime, config.Loaded.RefreshTokenCleanupInterval)

	// Initialize the Redis client
	options, err := redis.ParseURL(config.Loaded.RedisURL)
	if err != nil {
		logrus.WithError(err).Fatal()
	}
	options.OnConnect = func(_ context.Context, _ *redis.Conn) error {
		logrus.Info("Opening new Redis connection")
		return nil
	}
	rdb := redis.NewClient(options)
	defer func() {
		logrus.Info("Closing the Redis connection pool")
		if err := rdb.Close(); err != nil {
			logrus.WithError(err).Error()
		}
	}()

	// Set the pre-defined domains
	if err := setDomains(rdb, config.Loaded.DomainOverride); err != nil {
		logrus.WithError(err).Fatal()
	}

	// Initialize the karen logrus hook
	logrus.AddHook(&karen.LogrusHook{Redis: rdb})

	// Start up the REST API
	restApi := &api.API{
		Services: &api.Services{
			Accounts:      driver.Accounts,
			RefreshTokens: driver.RefreshTokens,
			Invites:       driver.Invites,
			Mailboxes:     driver.Mailboxes,
			Messages:      driver.Messages,
			Redis:         rdb,
		},
	}
	go func() {
		if err := restApi.Serve(); err != nil {
			logrus.WithError(err).Fatal()
		}
	}()
	defer func() {
		if err := restApi.Shutdown(); err != nil {
			logrus.WithError(err).Error()
		}
	}()

	// Notify karen about the service startup and shutdown
	if static.ApplicationMode == "PROD" {
		if err := karen.Send(rdb, karen.Message{
			Type:        karen.MessageTypeInfo,
			Service:     static.KarenServiceName,
			Topic:       "Startup",
			Description: "The service is now running.",
		}); err != nil {
			logrus.WithError(err).Error()
		}

		defer func() {
			if err := karen.Send(rdb, karen.Message{
				Type:        karen.MessageTypeInfo,
				Service:     static.KarenServiceName,
				Topic:       "Shutdown",
				Description: "The service has shut down.",
			}); err != nil {
				logrus.WithError(err).Error()
			}
		}()
	}

	// Wait for the program to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}

func refreshTokenCleanup(ctx context.Context, service shared.RefreshTokenService, lifetime, delay time.Duration) {
	logrus.Info("Starting the refresh token cleanup task")
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Shutting down the refresh token cleanup task")
			return
		case <-time.After(delay):
			deleted, err := service.DeleteExpired(lifetime)
			if err != nil {
				logrus.WithError(err).Error("Error while deleting expired refresh tokens")
				continue
			}
			logrus.Infof("Deleted %d expired refresh tokens", deleted)
		}
	}
}

func setDomains(rdb *redis.Client, domains []string) error {
	processed := make([]interface{}, len(domains))
	for i := range processed {
		processed[i] = domains[i]
	}
	if len(processed) > 0 {
		if err := rdb.Del(context.Background(), static.DomainsRedisKey).Err(); err != nil {
			return err
		}
		if err := rdb.SAdd(context.Background(), static.DomainsRedisKey, processed...).Err(); err != nil {
			return err
		}
	}
	return nil
}
