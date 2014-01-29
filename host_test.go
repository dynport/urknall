package urknall

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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

func TestInterfaceHandling(t *testing.T) {
	defaultInterface := "eth0"
	Convey("Given a host", t, func() {
		h := &Host{IP: "127.0.0.1"}
		Convey("When no interface is set", func() {
			Convey("And the Interface method is called", func() {
				v := h.publicInterface()
				Convey("Then the interface is set to the default", func() {
					So(v, ShouldEqual, defaultInterface)
				})
			})
		})

		Convey("When adding packages", func() {
			So(func() { h.AddPackage("pkg", nil) }, ShouldNotPanic)
			So(func() { h.AddPackage("other_pkg", nil) }, ShouldNotPanic)
			Convey("when adding a package with the same name", func() {
				So(func() { h.AddPackage("pkg", nil) }, ShouldPanic)
			})
		})

		Convey("When the interface is explicitly set to the default", func() {
			h.Interface = defaultInterface
			Convey("And the Interface method is called", func() {
				v := h.publicInterface()
				Convey("Then the interface is set to the default", func() {
					So(v, ShouldEqual, defaultInterface)
				})
			})
		})

		Convey("When the interface is set to 'tun0'", func() {
			h.Interface = "tun0"
			Convey("And the Interface method is called", func() {
				v := h.publicInterface()
				Convey("Then the interface is set to 'tun0'", func() {
					So(v, ShouldEqual, "tun0")
				})
			})
		})
	})
}
