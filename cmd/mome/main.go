package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/karta0898098/mome/pkg/configs"
	"github.com/karta0898098/mome/pkg/logging"
)

var (
	configPath string
)

var (
	CommitID = "0000000"
	Version  = "NULL_VERSION"
)

func init() {
	flag.StringVar(&configPath, "c", "", "--c {config path}")
}

func main() {
	// new context to control all application lifetimes
	// like gRPC server or background worker etc ...
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// for wait all goroutines
	wg := &sync.WaitGroup{}

	configProvider, err := configs.NewConfig(configPath)
	if err != nil {
		log.Fatal().Msgf("failed to new configuration %v", err)
	}

	logger := logging.Setup(configProvider.Get().Log)

	// using otlp inject trace id
	InitializesLocalOTLPProvider()

	app := NewApplication(configProvider, logger)

	go app.startReceiveTrade(ctx, wg)
	go app.startGRPCServer(ctx, wg)

	// wait close signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// send context done event
	// let all workers or server graceful shutdown
	cancel()
	wg.Wait()
}
