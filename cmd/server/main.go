package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/poopmail/canalization/internal/api"
	"github.com/poopmail/canalization/internal/database/postgres"
	"github.com/poopmail/canalization/internal/env"
	"github.com/poopmail/canalization/internal/static"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load the optional .env file
	godotenv.Load()

	// Initialize the postgres database driver
	driver, err := postgres.NewDriver(env.MustString("CANAL_POSTGRES_DSN", ""))
	if err != nil {
		log.Fatalln(err)
	}
	if err := driver.Migrate(); err != nil {
		log.Fatalln(err)
	}

	// Connect to the configured redis host
	rdb := redis.NewClient(&redis.Options{
		Addr: env.MustString("CANAL_REDIS_ADDRESS", "localhost:6379"),
		OnConnect: func(_ context.Context, _ *redis.Conn) error {
			logrus.Info("Opening new Redis connection")
			return nil
		},
		Username: env.MustString("CANAL_REDIS_USERNAME", ""),
		Password: env.MustString("CANAL_REDIS_PASSWORD", ""),
		DB:       0,
	})

	// Set the pre-defined domains
	rawDomains := env.MustStringSlice("CANAL_DOMAIN_OVERRIDE", ",", []string{})
	domains := make([]interface{}, 0, len(rawDomains))
	for _, rawDomain := range rawDomains {
		domains = append(domains, rawDomain)
	}
	if len(domains) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := rdb.Del(ctx, "__domains").Err(); err != nil {
			log.Fatal(err)
		}

		if err := rdb.SAdd(ctx, "__domains", domains...).Err(); err != nil {
			log.Fatal(err)
		}
	}

	// Start up the REST API
	restApi := &api.API{
		Settings: &api.Settings{
			Address:           env.MustString("CANAL_API_ADDRESS", ":8080"),
			RequestsPerMinute: env.MustInt("CANAL_API_RATE_LIMIT", 60),
			Production:        static.ApplicationMode == "PROD",
			Version:           static.ApplicationVersion,
		},
		Services: &api.Services{
			Invites:   driver.Invites,
			Accounts:  driver.Accounts,
			Mailboxes: driver.Mailboxes,
			Messages:  driver.Messages,
			Redis:     rdb,
		},
	}
	go func() {
		if err := restApi.Serve(); err != nil {
			logrus.Fatal(err)
		}
	}()
	defer func() {
		if err := restApi.Shutdown(); err != nil {
			logrus.Error(err)
		}
	}()

	// Wait for the program to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
