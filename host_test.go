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

		Convey("When the interface is explicitly set to the default", func() {
			h.SetInterface(defaultInterface)
			Convey("And the Interface method is called", func() {
				v := h.publicInterface()
				Convey("Then the interface is set to the default", func() {
					So(v, ShouldEqual, defaultInterface)
				})
			})
		})

		Convey("When the interface is set to 'tun0'", func() {
			h.SetInterface("tun0")
			Convey("And the Interface method is called", func() {
				v := h.publicInterface()
				Convey("Then the interface is set to 'tun0'", func() {
					So(v, ShouldEqual, "tun0")
				})
			})
		})
	})
}

func TestDockerConfiguration(t *testing.T) {
	defaultVersion := "0.7.0"
	Convey("Given a host", t, func() {
		h := &Host{IP: "127.0.0.1"}
		Convey("Given no docker configuration is set", func() {
			Convey("Then the docker host predicate should return 'false'", func() {
				So(h.isDockerHost(), ShouldBeFalse)
			})
			Convey("Then the docker build host predicate should return 'false'", func() {
				So(h.isDockerBuildHost(), ShouldBeFalse)
			})
			Convey("When the docker version is requested", func() {
				f := func() { h.dockerVersion() }
				Convey("Then a panic occurs", func() {
					So(f, ShouldPanicWith, "not a docker host")
				})
			})
		})

		Convey("Given the default docker configuration", func() {
			h.Docker = &DockerSettings{}
			Convey("Then the docker host predicate should return 'true'", func() {
				So(h.isDockerHost(), ShouldBeTrue)
			})
			Convey("Then the docker build host predicate should return 'false'", func() {
				So(h.isDockerBuildHost(), ShouldBeFalse)
			})
			Convey("When the docker version is requested", func() {
				v := h.dockerVersion()
				Convey("Then the default is returned", func() {
					So(v, ShouldEqual, defaultVersion)
				})
			})
		})

		Convey("Given the docker configuration with the docker build host flag set", func() {
			h.Docker = &DockerSettings{WithBuildSupport: true}
			Convey("Then the docker build host predicate should return 'true'", func() {
				So(h.isDockerBuildHost(), ShouldBeTrue)
			})
		})

		Convey("Given the docker configuration with a docker version explicitly set", func() {
			version := "0.5.3"
			h.Docker = &DockerSettings{Version: version}
			Convey("When the docker version is requested", func() {
				v := h.dockerVersion()
				Convey("Then the given version is returned", func() {
					So(v, ShouldEqual, version)
				})
			})
		})
	})
}
