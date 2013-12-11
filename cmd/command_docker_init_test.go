package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDockerInitCommand(t *testing.T) {
	Convey("Given a docker init command", t, func() {
		diC := &DockerInitCommand{Command: "do crazy shit!"}

		Convey("When the Command method Docker is called", func() {
			v := diC.Docker()
			Convey("Then a proper docker start command is returned", func() {
				So(v, ShouldEqual, "CMD do crazy shit!")
			})
		})

		Convey("When the Command method Shell is called", func() {
			v := diC.Shell()
			Convey("Then nothing is returned (wouldn't work for a regular host anyway)", func() {
				So(v, ShouldEqual, "")
			})
		})
		Convey("When the Command method Logging is called", func() {
			v := diC.Logging()
			Convey("Then a nice documentation string is returned", func() {
				So(v, ShouldEqual, "[D.RUN  ] Adding docker init cmd: do crazy shit!")
			})
		})
	})
}
