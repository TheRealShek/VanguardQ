package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheRealShek/VanguardQ/internal/queue"
	"github.com/TheRealShek/VanguardQ/internal/server"
)

func main() {
	go queue.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	// os.Interrupt — triggered by Ctrl+C in terminal or
	// syscall.SIGTERM — triggered by kill <pid> or Docker/Air stopping your app
	go server.Start()

	<-quit // blocks until Ctrl+C or kill signal
	fmt.Println("Shutting down...")
}
