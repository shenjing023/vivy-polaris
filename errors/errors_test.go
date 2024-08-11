package errors

import (
	"testing"
	"time"

	"errors"

	erro "github.com/cockroachdb/errors"

	_ "unsafe"
)

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
