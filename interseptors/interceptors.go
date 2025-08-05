package interceptor

import (
	"context"
	"github.com/kkiling/goplatform/log"
	"github.com/kkiling/goplatform/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewPanicRecoverInterceptor интерспетор паники
func NewPanicRecoverInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("panic recovered: %v", r)
				err = server.ErrInternal(err)
			}
		}()

		resp, err = handler(ctx, req)
		return resp, err
	}
}

// NewLoggerInterceptor интерспетор логирования
func NewLoggerInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if err == nil {
			return resp, err
		}
		switch status.Code(err) {
		case codes.Internal:
			logger.Errorf(err.Error())
		default:
			logger.Warnf(err.Error())
		}
		return resp, err
	}
}
