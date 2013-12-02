package zwo

import (
	"github.com/dynport/zwo/host"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCommandActionDocker(t *testing.T) {
	Convey("Given a command", t, func() {
		rawCmd := "do something"
		Convey("When the command is converted for docker", func() {
			h, _ := host.New("127.0.0.1", "", "")
			c := &commandAction{cmd: rawCmd, host: h}
			v := c.Docker()
			Convey("The value should be a valid dockerfile line", func() {
				So(v, ShouldEqual, "RUN do something")
			})
		})
	})
}

func TestCommandActionLogging(t *testing.T) {
	Convey("Given a command", t, func() {
		rawCmd := "do something"
		Convey("When the command is converted for logging", func() {
			h, _ := host.New("127.0.0.1", "", "")
			c := &commandAction{cmd: rawCmd, host: h}
			Convey("When the command has no further settings", func() {
				Convey("The line should have the COMMAND hint and contain the command", func() {
					v := c.Logging()
					So(v, ShouldEqual, "[COMMAND] # do something")
				})
			})

			Convey("When the host user is set to something other then 'root'", func() {
				c.host, _ = host.New("127.0.0.1", "gfrey", "")
				Convey("Then the SUDO hint should be added to the logging line", func() {
					v := c.Logging()
					So(v, ShouldContainSubstring, "[SUDO]")
				})
			})

			Convey("When the command user is set", func() {
				c.user = "gfrey"
				Convey("Then the SU hint should be added to the logging line with the correct user", func() {
					v := c.Logging()
					So(v, ShouldContainSubstring, "[SU:gfrey]")
				})
			})
		})
	})
}

func TestCommandActionShell(t *testing.T) {
	Convey("Given a command", t, func() {
		rawCmd := "do something"
		Convey("When the command is converted for shell", func() {
			h, _ := host.New("127.0.0.1", "", "")
			c := &commandAction{cmd: rawCmd, host: h}
			Convey("Then without further settings the value should be equal to the source", func() {
				v := c.Shell()
				So(v, ShouldEqual, rawCmd)
			})

			Convey("When the command should be run as root (aka no user given)", func() {
				Convey("When the user for the host is not given", func() {
					c.host, _ = host.New("127.0.0.1", "", "")
					v := c.Shell()
					Convey("The value should be a valid simple bash command", func() {
						So(v, ShouldEqual, rawCmd)
					})
				})

				Convey("When the user for the host is explicitly set to 'root'", func() {
					c.host, _ = host.New("127.0.0.1", "root", "")
					v := c.Shell()
					Convey("The value should be valid simple bash command", func() {
						So(v, ShouldEqual, rawCmd)
					})
				})

				Convey("When the user for the host is set to something different then 'root'", func() {
					c.host, _ = host.New("127.0.0.1", "gfrey", "")
					v := c.Shell()
					Convey("The value should be valid sudo command", func() {
						So(v, ShouldStartWith, "sudo bash <<EOF")
						So(v, ShouldContainSubstring, rawCmd)
					})
				})
			})

			Convey("When command should be run as a different user", func() {
				c.user = "gfrey"

				Convey("When the user for the host is not given", func() {
					c.host, _ = host.New("127.0.0.1", "", "")
					v := c.Shell()
					Convey("The value should be a valid bash command run as user 'gfrey'", func() {
						So(v, ShouldStartWith, "su -l gfrey <<EOF\n")
						So(v, ShouldContainSubstring, rawCmd)
					})
				})

				Convey("When the user for the host is explicitly set to 'root'", func() {
					c.host, _ = host.New("127.0.0.1", "root", "")
					v := c.Shell()
					Convey("The value should be a valid bash command run as user 'gfrey'", func() {
						So(v, ShouldStartWith, "su -l gfrey <<EOF\n")
						So(v, ShouldContainSubstring, rawCmd)
					})
				})

				Convey("When the user for the host is set to something different then 'root'", func() {
					c.host, _ = host.New("127.0.0.1", "gfrey", "")
					v := c.Shell()
					Convey("The value should be a valid bash command run with sudo and as user 'gfrey'", func() {
						So(v, ShouldStartWith, "sudo -- su -l gfrey <<EOF")
						So(v, ShouldContainSubstring, rawCmd)
					})
				})
			})
		})
	})
}
