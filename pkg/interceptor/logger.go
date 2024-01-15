package interceptor

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	logger = log.Logger
}

var (
	logger         zerolog.Logger
	loggableEvents []LoggableEvent
)

func SetLogger(l zerolog.Logger) {
	logger = l
}

func SetLoggableEvents(events []LoggableEvent) {
	loggableEvents = events
}

type ServerLoggingOption interface {
	apply(*serverLoggingOptions)
}

type LoggableEvent uint

const (
	// AccessLog is a loggable event representing start of the gRPC call.
	AccessLog LoggableEvent = iota
	RequestDump
	ResponseDump
)

func has(events []LoggableEvent, event LoggableEvent) bool {
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}

// serverLoggingOptions grpc client dump option
type serverLoggingOptions struct {
	Skipper        []func(req any, info *grpc.UnaryServerInfo) bool
	LoggableEvents []LoggableEvent
}

type serverLoggingSkipper struct {
	Skipper []func(req any, info *grpc.UnaryServerInfo) bool
}

func (s *serverLoggingSkipper) apply(options *serverLoggingOptions) {
	options.Skipper = s.Skipper
}

// WithSkipper grpc logging for skip some logs
func WithSkipper(skips ...func(req any, info *grpc.UnaryServerInfo) bool) ServerLoggingOption {
	return &serverLoggingSkipper{
		Skipper: skips,
	}
}

type serverLoggingLoggableEvents struct {
	LoggableEvents []LoggableEvent
}

func (s *serverLoggingLoggableEvents) apply(options *serverLoggingOptions) {
	options.LoggableEvents = s.LoggableEvents
}

func WithLoggableEvents(loggableEvents ...LoggableEvent) ServerLoggingOption {
	SetLoggableEvents(loggableEvents)
	return &serverLoggingLoggableEvents{
		LoggableEvents: loggableEvents,
	}
}

// UnaryServerLoggerInterceptor is spawn MyLogger to each request context
func UnaryServerLoggerInterceptor(baseLogger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		l := baseLogger
		sp := trace.SpanFromContext(ctx).SpanContext()
		if sp.HasTraceID() && sp.HasSpanID() {
			l = l.With().
				Str("traceId", sp.TraceID().String()).
				Str("spanId", sp.SpanID().String()).Logger()
		}

		ctx = l.WithContext(ctx)
		resp, err = handler(ctx, req)
		return
	}
}

// UnaryServerLoggingInterceptor logging grpc access log and reply status
func UnaryServerLoggingInterceptor(opts ...ServerLoggingOption) grpc.UnaryServerInterceptor {
	opt := &serverLoggingOptions{
		Skipper: nil,
	}

	for _, o := range opts {
		o.apply(opt)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		startTime := time.Now()

		logger := log.Ctx(ctx)

		skipLogging := false
		for _, skip := range opt.Skipper {
			if skip != nil && skip(req, info) {
				skipLogging = true
				break
			}
		}

		if has(loggableEvents, RequestDump) && !skipLogging {
			requestDump(ctx, info, logger, req, err)
		}

		resp, err = handler(ctx, req)

		if has(loggableEvents, AccessLog) && !skipLogging {
			logger.WithLevel(DefaultServerCodeToLevel(status.Code(err))).
				Str("method", info.FullMethod).
				Uint32("code", uint32(status.Code(err))).
				Dur("since", time.Since(startTime)).
				Msg("grpc request access log")
		}

		if has(loggableEvents, ResponseDump) && !skipLogging {
			replayDump(ctx, info, logger, resp, err)
		}

		return
	}
}

// DefaultServerCodeToLevel is the helper mapper that maps gRPC return codes to log levels for server side.
func DefaultServerCodeToLevel(code codes.Code) zerolog.Level {
	switch code {
	case codes.OK, codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated:
		return zerolog.InfoLevel

	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted,
		codes.OutOfRange, codes.Unavailable:
		return zerolog.WarnLevel

	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return zerolog.ErrorLevel

	default:
		return zerolog.InfoLevel
	}
}

// ClientLoggingOption grpc client logging
type ClientLoggingOption interface {
	apply(*clientLoggingOptions)
}

// clientLoggingOptions grpc client dump option
type clientLoggingOptions struct {
	Dump bool
}

// apply is implement ClientLoggingOption
func (c *clientLoggingOptions) apply(opts *clientLoggingOptions) {
	opts.Dump = c.Dump
}

// WithClientDump grpc logging for enable dump req/reply log
// useful to help debug
func WithClientDump(dump bool) ClientLoggingOption {
	return &clientLoggingOptions{
		Dump: dump,
	}
}

// UnaryClientLoggingInterceptor logging grpc invoke grpc log
func UnaryClientLoggingInterceptor(opts ...ClientLoggingOption) grpc.UnaryClientInterceptor {
	opt := &clientLoggingOptions{
		Dump: false,
	}

	for _, o := range opts {
		o.apply(opt)
	}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		logger := log.Ctx(ctx)

		l := logger.WithLevel(DefaultServerCodeToLevel(status.Code(err)))

		if opt.Dump {
			l = l.Interface("req", req).Interface("reply", reply)
		}

		l.
			Str("method", method).
			Msg("invoke grpc log")

		return err
	}
}
