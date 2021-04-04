package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// Start up the REST API
	restApi := &api.API{
		Settings: &api.Settings{
			Address:           env.MustString("CANAL_API_ADDRESS", ":8080"),
			RequestsPerMinute: env.MustInt("CANAL_API_RATE_LIMIT", 60),
			Production:        static.ApplicationMode == "PROD",
			Version:           static.ApplicationVersion,
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
