package utils

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseVersion(t *testing.T) {
	Convey("ParseVersion", t, func() {
		Convey("Parse Version", func() {
			v := &Version{}
			So(v.Parse("1.2.3"), ShouldBeNil)
			So(v.Major, ShouldEqual, 1)
			So(v.Minor, ShouldEqual, 2)
			So(v.Patch, ShouldEqual, 3)
		})

		a, e := ParseVersion("0.1.2")
		So(e, ShouldBeNil)

		Convey("Compare Smaller Versions", func() {
			for _, v := range []string{"0.1.3", "0.2.0", "1.0.0"} {
				b, e := ParseVersion(v)
				So(e, ShouldBeNil)
				So(a.Smaller(b), ShouldBeTrue)
			}
		})

		Convey("Compare Equal Version", func() {
			b, e := ParseVersion("0.1.2")
			So(e, ShouldBeNil)
			So(a.Smaller(b), ShouldBeFalse)
		})
	})
}
