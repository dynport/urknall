package cmd

import (
	"github.com/dynport/zwo/host"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDockerInitCommand(t *testing.T) {
	h, _ := host.New("127.0.0.1", "", "")
	Convey("Given a docker init command", t, func() {
		diC := &DockerInitCommand{Command: "do crazy shit!"}

		Convey("When the Commander method Docker is called", func() {
			v := diC.Docker(h)
			Convey("Then a proper docker start command is returned", func() {
				So(v, ShouldEqual, "CMD do crazy shit!")
			})
		})

		Convey("When the Commander method Shell is called", func() {
			v := diC.Shell(h)
			Convey("Then nothing is returned (wouldn't work for a regular host anyway)", func() {
				So(v, ShouldEqual, "")
			})
		})
		Convey("When the Commander method Logging is called", func() {
			v := diC.Logging(h)
			Convey("Then a nice documentation string is returned", func() {
				So(v, ShouldEqual, "[D.RUN  ] Adding docker init cmd: do crazy shit!")
			})
		})
	})
}
