package interceptor

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerErrorHandleInterceptor is logging error
func UnaryServerErrorHandleInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			st := status.Convert(err)
			logger := log.Ctx(ctx)
			logger.
				WithLevel(DefaultServerCodeToLevel(st.Code())).
				Str("method", info.FullMethod).
				Msg("grpc occur error")
		}
		return
	}
}
