package errors

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	_ "unsafe"

	"bou.ke/monkey"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	Code  = codes.Code
	Error struct {
		Code Code
		Err  error
	}
)

//go:linkname strToCode google.golang.org/grpc/codes.strToCode
var strToCode map[string]codes.Code

const (
	InternalError = "Service Internal Error"
)

const (
	OK Code = iota
	Canceled
	Unknown
	InvalidArgument
	DeadlineExceeded
	NotFound
	AlreadyExists
	PermissionDenied
	ResourceExhausted
	FailedPrecondition
	Aborted
	OutOfRange
	Unimplemented
	Internal
	Unavailable
	DataLoss
	Unauthenticated
)

var (
	// ErrMap 对应的grpc error code
	errMap = map[Code]codes.Code{
		OK:                 codes.OK,
		Canceled:           codes.Canceled,
		Unknown:            codes.Unknown,
		InvalidArgument:    codes.InvalidArgument,
		DeadlineExceeded:   codes.DeadlineExceeded,
		NotFound:           codes.NotFound,
		AlreadyExists:      codes.AlreadyExists,
		PermissionDenied:   codes.PermissionDenied,
		ResourceExhausted:  codes.ResourceExhausted,
		FailedPrecondition: codes.FailedPrecondition,
		Aborted:            codes.Aborted,
		OutOfRange:         codes.OutOfRange,
		Unimplemented:      codes.Unimplemented,
		Internal:           codes.Internal,
		Unavailable:        codes.Unavailable,
		DataLoss:           codes.DataLoss,
		Unauthenticated:    codes.Unauthenticated,
	}
)

func init() {
	var a codes.Code
	monkey.PatchInstanceMethod(reflect.TypeOf(a), "String", func(c codes.Code) string {
		for k, v := range strToCode {
			if v == c {
				return strings.ReplaceAll(k, `"`, "")
			}
		}
		return "Code(" + strconv.FormatInt(int64(c), 10) + ")"
	})
	/*
		monkey.PatchInstanceMethod(reflect.TypeOf(&a), "UnmarshalJSON", func(c *codes.Code, b []byte) error {
			// From json.Unmarshaler: By convention, to approximate the behavior of
			// Unmarshal itself, Unmarshalers implement UnmarshalJSON([]byte("null")) as
			// a no-op.
			if string(b) == "null" {
				return nil
			}
			fmt.Println(string(b))
			if c == nil {
				return fmt.Errorf("nil receiver passed to UnmarshalJSON")
			}

			if ci, err := strconv.ParseUint(string(b), 10, 32); err == nil {
				if ci >= 17 {
					return fmt.Errorf("invalid code: %q", ci)
				}

				*c = Code(ci)
				return nil
			}

			fmt.Printf("%+v\n", strToCode)
			if jc, ok := strToCode[string(b)]; ok {
				*c = jc
				return nil
			}
			fmt.Println("sadadada" + string(b))
			return fmt.Errorf("invalid code111: %q", string(b))
		})
	*/
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %+v", e.Code, e.Err)
}

func (e *Error) Format(s fmt.State, verb rune) {
	errors.FormatError(e, s, verb)
}

// if don't want print stack message, use this
// func (e *Error) SafeFormatError(p errors.Printer) (next error) {
// 	if p.Detail() {
// 		p.Printf("code[%s]: %s", errors.Safe(e.Code), errors.Safe(e.Err.Error()))
// 	}
// 	return e.Err
// }

// ServiceErr2GRPCErr serviceErr covert to GRPCErr
func ServiceErr2GRPCErr(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		if e.Code == Internal {
			return status.Error(errMap[Internal], "Service Internal Error")
		}
		if _, ok := errMap[e.Code]; ok {
			return status.Error(errMap[e.Code], e.Err.Error())
		}
		return status.Error(codes.Unknown, e.Err.Error())
	}
	return status.Error(codes.Unknown, err.Error())
}

func NewServiceErr(code Code, err error) *Error {
	return &Error{code, err}
}

// CodePair custom error code
type CodePair struct {
	Code Code
	// Desc is the error brief description with upper case, like "CUSTOM_ERROR"
	Desc string
}

// RegisterErrCode append custom error code
func RegisterErrCode(pairs ...CodePair) {
	for _, pair := range pairs {
		errMap[pair.Code] = pair.Code
		strToCode[fmt.Sprintf(`"%s"`, pair.Desc)] = pair.Code
	}
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
