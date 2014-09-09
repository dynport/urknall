package urknall

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testCommandCustomChecksum struct {
	*testCommand
}

func (c *testCommandCustomChecksum) Checksum() string {
	return "default checksum"
}

func TestUtils(t *testing.T) {
	c := &testCommand{}
	Convey("Command checksum", t, func() {
		Convey("for default commands", func() {
			checksum, e := commandChecksum(c)
			So(e, ShouldBeNil)
			So(checksum, ShouldEqual, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		})
		Convey("for custom checksums", func() {
			c := &testCommandCustomChecksum{c}
			checksum, e := commandChecksum(c)
			So(e, ShouldBeNil)
			So(checksum, ShouldEqual, "default checksum")
		})
	})
}
