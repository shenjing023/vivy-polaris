package common

import (
	er "github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/example/pb"
	"google.golang.org/grpc/codes"
)

const (
	CUSTOM_ERR_CODE1 er.Code = 100
)

var (
	CodeMap = map[er.Code]string{
		CUSTOM_ERR_CODE1: "CUSTOM_ERROR1",
	}
)

func init() {
	// for k, v := range CodeMap {
	// 	er.RegisterErrCode(er.CodePair{Code: k, Desc: v})
	// }
	for k, v := range pb.Code_name {
		er.RegisterErrCode(er.CodePair{Code: codes.Code(k), Desc: v})
	}
}
