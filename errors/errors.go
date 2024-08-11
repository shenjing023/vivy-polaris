package errors

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Code codes.Code
	Err  error
}

const (
	InternalError = "service internal error"
)

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %+v", e.Code, e.Err)
}

func (e *Error) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// ServiceErr2GRPCErr serviceErr covert to GRPCErr
func ServiceErr2GRPCErr(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := errors.Cause(err).(*Error); ok {
		return status.Error(e.Code, e.Err.Error())
	}
	return status.Error(codes.Unknown, err.Error())
}

func NewServiceErr(code codes.Code, err error) *Error {
	return &Error{code, err}
}

func NewInternalError() *Error {
	return &Error{codes.Internal, fmt.Errorf(InternalError)}
}

// GRPCErr2GQLErr grpc error convert to gql_service error
// func GRPCErr2GQLErr(ctx context.Context, err error, conf map[Code]string) error {
// 	st, ok := status.FromError(err)
// 	if !ok {
// 		// Error was not a status error
// 		return &gqlerror.Error{
// 			Path:    graphql.GetPath(ctx),
// 			Message: InternalError,
// 			Extensions: map[string]interface{}{
// 				"code": Internal,
// 			},
// 		}
// 	}
// 	var (
// 		errMsg = st.Message()
// 		code   = Unknown
// 	)
// 	for k, v := range conf {
// 		if k == st.Code() {
// 			errMsg = v
// 			code = errMap[k]
// 			break
// 		}
// 	}
// 	return &gqlerror.Error{
// 		Message: errMsg,
// 		Extensions: map[string]interface{}{
// 			"code": code,
// 		},
// 	}
// }

// ServerErrorInterceptor transfer a error to status error
func ServerErrorInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	return resp, ServiceErr2GRPCErr(err)
}

// func ClientErrorInterceptor(ctx context.Context, method string, req, reply interface{},
// 	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
// 	err := invoker(ctx, method, req, reply, cc, opts...)
// 	if err == nil {
// 		return nil
// 	}
// 	cause := errors.Cause(err)
// 	st, ok := status.FromError(cause)
// 	if ok {
// 		// details := st.Details()

// 	}
// 	return err
// }

func Convert(err error) *status.Status {
	s, _ := status.FromError(err)
	return s
}
