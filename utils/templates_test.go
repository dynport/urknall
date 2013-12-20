package utils

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMustRender(t *testing.T) {
	Convey("MustRenderTemplate", t, func(){
		type typ struct {
			Version string
			Path string
		}
		ins := typ{ Version: "1.2.3", Path: "/path/to/{{ .Version }}"}
		So(MustRenderTemplate("{{ .Path }}", ins), ShouldEqual, "/path/to/1.2.3")

	})
}
