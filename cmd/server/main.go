package main

import (
	"github.com/joho/godotenv"
	"github.com/poopmail/canalization/internal/database/postgres"
	"github.com/poopmail/canalization/internal/env"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// TODO: Implement startup logic

	// Wait for the program to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
