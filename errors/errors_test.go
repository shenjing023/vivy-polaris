package errors

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"errors"

	erro "github.com/cockroachdb/errors"

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

func TestWrap(t *testing.T) {
	startTime := time.Now()
	t.Logf("%+v", b())
	t.Logf("cost:%s", time.Since(startTime))
	// fmt.Printf("%+v \n", a())
	// fmt.Printf("%+v \n", b())
}

func a() error {
	err := c()
	return errors.New("c error: " + err.Error())
}

func b() error {
	err := c()
	return NewServiceErr(0, err)
}

func c() error {
	return erro.New("c")
}
