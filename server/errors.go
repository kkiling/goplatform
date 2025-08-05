package server

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrInternal .
	ErrInternal = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.Internal, err, details...)
	}
	// ErrNotFound .
	ErrNotFound = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.NotFound, err)
	}
	// ErrPermissionDenied .
	ErrPermissionDenied = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.PermissionDenied, err, details...)
	}
	// ErrInvalidArgument .
	ErrInvalidArgument = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.InvalidArgument, err, details...)
	}
	// ErrUnauthenticated .
	ErrUnauthenticated = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.Unauthenticated, err, details...)
	}
	// ErrAlreadyExists .
	ErrAlreadyExists = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.AlreadyExists, err, details...)
	}
	// ErrTooManyRequests .
	ErrTooManyRequests = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.ResourceExhausted, err, details...)
	}
	// ErrFailedPrecondition .
	ErrFailedPrecondition = func(err error, details ...proto.Message) error {
		return NewResponseError(codes.FailedPrecondition, err, details...)
	}
)

// NewResponseError .
func NewResponseError(code codes.Code, err error, details ...proto.Message) error {
	st, err := status.New(code, err.Error()).WithDetails(details...)

	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}

	return st.Err()
}
