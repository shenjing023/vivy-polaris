package common

import (
	er "github.com/shenjing023/vivy-polaris/errors"
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
	for k, v := range CodeMap {
		er.RegisterErrCode(er.CodePair{Code: k, Desc: v})
	}
}
