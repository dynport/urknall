package host

import (
	"github.com/dynport/zwo/firewall"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIPAddressHandling(t *testing.T) {
	Convey("Given a host with empty IP address", t, func() {
		ip := ""
		Convey("When the New function is called", func() {
			h, e := New(ip, "", "")
			Convey("Then an error is returned", func() {
				So(e.Error(), ShouldEqual, "no IP address given")
			})
			Convey("Then the host value returned is nil", func() {
				So(h, ShouldBeNil)
			})
		})
	})

	Convey("Given a host with an invalid IP address", t, func() {
		ip := "not an ip address"
		Convey("When the New function is called", func() {
			h, e := New(ip, "", "")
			Convey("Then an error is returned", func() {
				So(e.Error(), ShouldEqual, "not a valid IP address (must be either IPv4 or IPv6): not an ip address")
			})
			Convey("Then the host value returned is nil", func() {
				So(h, ShouldBeNil)
			})
		})
	})

	Convey("Given a host with something like an IP address (but not one though)", t, func() {
		ip := "666.666.666.666"
		Convey("When the New function is called", func() {
			h, e := New(ip, "", "")
			Convey("Then an error is returned", func() {
				So(e.Error(), ShouldEqual, "not a valid IP address (must be either IPv4 or IPv6): 666.666.666.666")
			})
			Convey("Then the host value returned is nil", func() {
				So(h, ShouldBeNil)
			})
		})
	})

	Convey("Given a host with an valid IPv4 address", t, func() {
		ip := "127.0.0.1"
		Convey("When the New function is called", func() {
			h, e := New(ip, "", "")
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's IP address equals the specified one", func() {
				So(h.IPAddress(), ShouldEqual, ip)
			})
			Convey("Then the error returned is nil", func() {
				So(e, ShouldBeNil)
			})
		})
	})

	Convey("Given a host with an valid IPv6 address", t, func() {
		ip := "abcd::1"
		Convey("When the New function is called", func() {
			h, e := New(ip, "", "")
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's IP address equals the specified one", func() {
				So(h.IPAddress(), ShouldEqual, ip)
			})
			Convey("Then the error returned is nil", func() {
				So(e, ShouldBeNil)
			})
		})
	})
}

func TestUserHandling(t *testing.T) {
	Convey("Given a host with no user specified", t, func() {
		user := ""
		Convey("When the New function is called", func() {
			h, e := New("127.0.0.1", user, "")
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals root", func() {
				So(h.User(), ShouldEqual, "root")
			})
			Convey("The returned host's 'isSudoRequired' predicate return false", func() {
				So(h.IsSudoRequired(), ShouldBeFalse)
			})
			Convey("Then the error returned is nil", func() {
				So(e, ShouldBeNil)
			})
		})
	})

	Convey("Given a host with user set to 'root'", t, func() {
		user := "root"
		Convey("When the New function is called", func() {
			h, e := New("127.0.0.1", user, "")
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals root", func() {
				So(h.User(), ShouldEqual, "root")
			})
			Convey("The returned host's 'isSudoRequired' predicate return false", func() {
				So(h.IsSudoRequired(), ShouldBeFalse)
			})
			Convey("Then the error returned is nil", func() {
				So(e, ShouldBeNil)
			})
		})
	})

	Convey("Given a host with user set to something not being 'root'", t, func() {
		user := "gfrey"
		Convey("When the New function is called", func() {
			h, e := New("127.0.0.1", user, "")
			Convey("Then a host value is returned", func() {
				So(h, ShouldNotBeNil)
			})
			Convey("The returned host's user equals the given one", func() {
				So(h.User(), ShouldEqual, user)
			})
			Convey("The returned host's 'isSudoRequired' predicate return true", func() {
				So(h.IsSudoRequired(), ShouldBeTrue)
			})
			Convey("Then the error returned is nil", func() {
				So(e, ShouldBeNil)
			})
		})
	})
}

func TestInterfaceHandling(t *testing.T) {
	defaultInterface := "eth0"
	Convey("Given a host", t, func() {
		h, _ := New("127.0.0.1", "", "")
		Convey("When no interface is set", func() {
			Convey("And the Interface method is called", func() {
				v := h.Interface()
				Convey("Then the interface is set to the default", func() {
					So(v, ShouldEqual, defaultInterface)
				})
			})
		})

		Convey("When the interface is explicitly set to the default", func() {
			h.SetInterface(defaultInterface)
			Convey("And the Interface method is called", func() {
				v := h.Interface()
				Convey("Then the interface is set to the default", func() {
					So(v, ShouldEqual, defaultInterface)
				})
			})
		})

		Convey("When the interface is set to 'tun0'", func() {
			h.SetInterface("tun0")
			Convey("And the Interface method is called", func() {
				v := h.Interface()
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
		h, _ := New("127.0.0.1", "", "")
		Convey("Given no docker configuration is set", func() {
			Convey("Then the docker host predicate should return 'false'", func() {
				So(h.IsDockerHost(), ShouldBeFalse)
			})
			Convey("Then the docker build host predicate should return 'false'", func() {
				So(h.IsDockerBuildHost(), ShouldBeFalse)
			})
			Convey("When the docker version is requested", func() {
				f := func() { h.DockerVersion() }
				Convey("Then a panic occurs", func() {
					So(f, ShouldPanicWith, "not a docker host")
				})
			})
		})

		Convey("Given the default docker configuration", func() {
			h.Docker = &DockerSettings{}
			Convey("Then the docker host predicate should return 'true'", func() {
				So(h.IsDockerHost(), ShouldBeTrue)
			})
			Convey("Then the docker build host predicate should return 'false'", func() {
				So(h.IsDockerBuildHost(), ShouldBeFalse)
			})
			Convey("When the docker version is requested", func() {
				v := h.DockerVersion()
				Convey("Then the default is returned", func() {
					So(v, ShouldEqual, defaultVersion)
				})
			})
		})

		Convey("Given the docker configuration with the docker build host flag set", func() {
			h.Docker = &DockerSettings{WithBuildSupport: true}
			Convey("Then the docker build host predicate should return 'true'", func() {
				So(h.IsDockerBuildHost(), ShouldBeTrue)
			})
		})

		Convey("Given the docker configuration with a docker version explicitly set", func() {
			version := "0.5.3"
			h.Docker = &DockerSettings{Version: version}
			Convey("When the docker version is requested", func() {
				v := h.DockerVersion()
				Convey("Then the given version is returned", func() {
					So(v, ShouldEqual, version)
				})
			})
		})
	})
}

func TestFirewallRuleHandling(t *testing.T) {
	Convey("Given a host", t, func() {
		h, _ := New("127.0.0.1", "", "")
		Convey("When no rules are set", func() {
			Convey("And they are retrieved", func() {
				r := h.FirewallRules()
				Convey("Then they are empty", func() {
					So(len(r), ShouldEqual, 0)
				})
			})
		})

		Convey("When a simple rule is set", func() {
			h.AddFirewallRule(firewall.DockerService("foo"))
			Convey("And they are retrieved", func() {
				r := h.FirewallRules()
				Convey("Then one rule is returned", func() {
					So(len(r), ShouldEqual, 1)
				})
			})
		})
	})
}
