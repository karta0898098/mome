package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/karta0898098/mome/pkg/logging"
)

func main() {
	logger := logging.SetupWithOption(
		logging.WithLevel(logging.DebugLevel),
		logging.WithDebug(true),
	)

	// new context to control all application lifetime
	// like gRPC server or background worker etc ...
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// for wait all goroutines
	wg := &sync.WaitGroup{}

	// wait close signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// send context done event
	// let all worker or server graceful shutdown
	cancel()
}
