package utils

import (
	"testing"
)

func TestMustRender(t *testing.T) {
	type typ struct {
		Version string
		Path    string
	}
	ins := typ{Version: "1.2.3", Path: "/path/to/{{ .Version }}"}
	res := MustRenderTemplate("{{ .Path }}", ins)
	if res != "/path/to/1.2.3" {
		t.Errorf("expected result to be %q, got %q", "/path/to/1.2.3", res)
	}
}
