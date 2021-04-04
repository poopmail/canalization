package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// TODO: Implement startup logic

	// Wait for the program to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
