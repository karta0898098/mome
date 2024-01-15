package main

import (
	"context"
	"net"
	"sync"

	pb "github.com/karta0898098/mome/pb/order"
	"github.com/karta0898098/mome/pkg/configs"
	"github.com/karta0898098/mome/pkg/interceptor"
	grpctransport "github.com/karta0898098/mome/pkg/transport/grpc"

	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

// Application contains this app need components
type Application struct {
	cfg     configs.ConfigurationProvider // Cfg is configuration provider. It provide all this application server need config.
	logger  zerolog.Logger                // application logger
	handler *grpctransport.OrderMatchingHandler
}

// startGRPCServer start grpc server
// ctx is control the all application lifecycle
// wg is for control all async job is finish or not
func (app *Application) startGRPCServer(ctx context.Context, wg *sync.WaitGroup) {
	var (
		server *grpc.Server
	)

	wg.Add(1)
	defer wg.Done()

	port := app.cfg.Get().GRPC.Port

	// start listen tcp port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		app.logger.Fatal().Err(err).Msgf("failed to listen on prot=%v", port)
	}

	// prepare the grpc interceptors
	interceptors := []grpc.UnaryServerInterceptor{
		interceptor.UnaryServerLoggerInterceptor(app.logger),
		interceptor.UnaryServerLoggingInterceptor(
			interceptor.WithLoggableEvents(app.cfg.Get().GRPC.LogEvents...)),
		interceptor.UnaryServerErrorHandleInterceptor(),
		interceptor.UnaryServerRecoveryInterceptor(),
	}

	server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	pb.RegisterOrderMatchingServiceServer(server, app.handler)
	reflection.Register(server)

	app.logger.Info().Msgf("start grpc server on %v", port)
	go func() {
		// service connections
		err = server.Serve(listener)
		if err != nil {
			app.logger.Error().Msgf("grpc serve : %s\n", err)
		}
	}()

	<-ctx.Done()

	// ignore error since it will be "Err shutting down server : context canceled"
	server.GracefulStop()

	app.logger.Info().Msgf("grpc server gracefully stopped")
}
