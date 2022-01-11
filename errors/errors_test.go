package errors

import (
	"encoding/json"
	"reflect"
	"testing"

	_ "unsafe"

	"bou.ke/monkey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMonkeyPatch(t *testing.T) {
	type MyError = codes.Code
	var got []MyError
	// want := []Code{OK, NotFound, Internal, Canceled}
	in := `["OK", "NOT_FOUND", "INTERNAL", "CANCELLED","sample"]`
	err := json.Unmarshal([]byte(in), &got)
	t.Logf("g: %v\n", got)
	t.Log(err)
	strToCode["OK1111"] = codes.OK
	t.Log(strToCode)

	var a codes.Code = 100
	monkey.PatchInstanceMethod(reflect.TypeOf(a), "String", func(_ codes.Code) string {
		return "OK112321"
	})
	t.Log(status.Error(100, "retry test"))
	t.Log(a.String())
}
