package configs

import "github.com/karta0898098/mome/pkg/interceptor"

// GRPCServer is define grpc server port
type GRPCServer struct {
	Port      string                      `mapstructure:"port"`
	LogEvents []interceptor.LoggableEvent `mapstructure:"logEvents"`
}
