package urknall

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUserHandling(t *testing.T) {
	Convey("Given a host with no user specified", t, func() {
		user := ""
		Convey("When the New function is called", func() {
			h := &Host{User: user}
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals root", func() {
				So(h.user(), ShouldEqual, "root")
			})
			Convey("The returned host's 'isSudoRequired' predicate return false", func() {
				So(h.isSudoRequired(), ShouldBeFalse)
			})
		})
	})

	Convey("Given a host with user set to 'root'", t, func() {
		user := "root"
		Convey("When the New function is called", func() {
			h := &Host{User: user}
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals root", func() {
				So(h.user(), ShouldEqual, "root")
			})
			Convey("The returned host's 'isSudoRequired' predicate return false", func() {
				So(h.isSudoRequired(), ShouldBeFalse)
			})
		})
	})

	Convey("Given a host with user set to something not being 'root'", t, func() {
		user := "gfrey"
		Convey("When the New function is called", func() {
			h := &Host{User: user}
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals the given one", func() {
				So(h.user(), ShouldEqual, user)
			})
			Convey("The returned host's 'isSudoRequired' predicate return true", func() {
				So(h.isSudoRequired(), ShouldBeTrue)
			})
		})
	})
}
