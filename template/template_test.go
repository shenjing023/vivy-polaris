package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPkgName(t *testing.T) {
	pkgName := ImportPathForDir(".")
	assert.Equal(t, "github.com/shenjing023/vivy-polaris/template", pkgName)
}

func TestRender(t *testing.T) {
	data := struct {
		PkgName string
	}{
		PkgName: ImportPathForDir("."),
	}
	err := Render(data)
	assert.Nil(t, err)
}
