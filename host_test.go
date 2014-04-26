package urknall

import (
	"testing"

	"github.com/dynport/urknall/ssh"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUserHandling(t *testing.T) {
	Convey("Given a host with no user specified", t, func() {
		Convey("When the New function is called", func() {
			h := &ssh.Host{Address: "127.0.0.1"}
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals root", func() {
				So(h.User(), ShouldEqual, "root")
			})
		})
	})

	Convey("Given a host with user set to 'root'", t, func() {
		Convey("When the New function is called", func() {
			h := &ssh.Host{Address: "root@127.0.0.1"}
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals root", func() {
				So(h.User(), ShouldEqual, "root")
			})
		})
	})

	Convey("Given a host with user set to something not being 'root'", t, func() {
		Convey("When the New function is called", func() {
			h := &ssh.Host{Address: "gfrey@127.0.0.1"}
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals the given one", func() {
				So(h.User(), ShouldEqual, "gfrey")
			})
		})
	})
}
