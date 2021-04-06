package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/poopmail/canalization/internal/api"
	"github.com/poopmail/canalization/internal/auth"
	"github.com/poopmail/canalization/internal/database/postgres"
	"github.com/poopmail/canalization/internal/env"
	"github.com/poopmail/canalization/internal/karen"
	"github.com/poopmail/canalization/internal/random"
	"github.com/poopmail/canalization/internal/static"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load the optional .env file
	godotenv.Load()

	// Initialize the postgres database driver
	driver, err := postgres.NewDriver(env.MustString("CANAL_POSTGRES_DSN", ""))
	if err != nil {
		logrus.WithError(err).Fatal()
	}
	if err := driver.Migrate(); err != nil {
		logrus.WithError(err).Fatal()
	}

	// Initialize the Redis client
	options, err := redis.ParseURL(env.MustString("CANAL_REDIS_URL", "redis://localhost:6379/0"))
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
	domains := env.MustStringSlice("CANAL_DOMAIN_OVERRIDE", ",", []string{})
	processed := make([]interface{}, len(domains))
	for i := range processed {
		processed[i] = domains[i]
	}
	if len(processed) > 0 {
		if err := rdb.Del(context.Background(), static.DomainsRedisKey).Err(); err != nil {
			logrus.WithError(err).Fatal()
		}
		if err := rdb.SAdd(context.Background(), static.DomainsRedisKey, processed...).Err(); err != nil {
			logrus.WithError(err).Fatal()
		}
	}

	// Initialize the karen logrus hook
	logrus.AddHook(&karen.LogrusHook{Redis: rdb})

	// Initialize the authenticator
	authenticator := auth.NewJWTAuthenticator(
		env.MustString("CANAL_AUTH_JWT_SIGNING_KEY", random.RandomString(64)),
		time.Duration(int64(env.MustInt("CANAL_AUTH_JWT_LIFETIME", 10080))*int64(time.Minute)),
		driver.Accounts,
	)

	// Start up the REST API
	restApi := &api.API{
		Settings: &api.Settings{
			Address:           env.MustString("CANAL_API_ADDRESS", ":8080"),
			RequestsPerMinute: env.MustInt("CANAL_API_RATE_LIMIT", 60),
			Production:        static.ApplicationMode == "PROD",
			Version:           static.ApplicationVersion,
		},
		Services: &api.Services{
			Authenticator: authenticator,
			Invites:       driver.Invites,
			Accounts:      driver.Accounts,
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
