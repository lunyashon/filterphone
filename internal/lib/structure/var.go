package structure

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var (
	Status = map[codes.Code]int{
		codes.OK:                 http.StatusOK,
		codes.InvalidArgument:    http.StatusBadRequest,
		codes.Internal:           http.StatusInternalServerError,
		codes.NotFound:           http.StatusNotFound,
		codes.Unauthenticated:    http.StatusUnauthorized,
		codes.PermissionDenied:   http.StatusForbidden,
		codes.Unimplemented:      http.StatusNotImplemented,
		codes.Unavailable:        http.StatusServiceUnavailable,
		codes.DeadlineExceeded:   http.StatusGatewayTimeout,
		codes.ResourceExhausted:  http.StatusTooManyRequests,
		codes.FailedPrecondition: http.StatusPreconditionFailed,
	}
)
