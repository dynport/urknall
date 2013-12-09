package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCommandActionDocker(t *testing.T) {
	Convey("Given a command", t, func() {
		rawCmd := "do something"
		Convey("When the command is converted for docker", func() {
			c := &ShellCommand{Command: rawCmd}
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
			c := &ShellCommand{Command: rawCmd}
			Convey("When the command has no further settings", func() {
				Convey("The line should have the COMMAND hint and contain the command", func() {
					v := c.Logging()
					So(v, ShouldEqual, "[COMMAND] # do something")
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
			c := &ShellCommand{Command: rawCmd}
			Convey("Then without further settings the value should be equal to the source", func() {
				v := c.Shell()
				So(v, ShouldEqual, rawCmd)
			})

			Convey("When the command should be run as root (aka no user given)", func() {
				Convey("When the user for the host is not given", func() {
					v := c.Shell()
					Convey("The value should be a valid simple bash command", func() {
						So(v, ShouldEqual, rawCmd)
					})
				})

				Convey("When the user for the host is explicitly set to 'root'", func() {
					v := c.Shell()
					Convey("The value should be valid simple bash command", func() {
						So(v, ShouldEqual, rawCmd)
					})
				})
			})

			Convey("When command should be run as a different user", func() {
				c.user = "gfrey"

				v := c.Shell()
				Convey("The value should be a valid bash command run as user 'gfrey'", func() {
					So(v, ShouldStartWith, "su -l gfrey <<EOF")
					So(v, ShouldContainSubstring, rawCmd)
				})
			})
		})
	})
}
