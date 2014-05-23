package urknall

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTaskImpl(t *testing.T) {
	Convey("Task Impl", t, func() {
		reference := struct{ Version string }{"1.2"}
		i := &task{taskBuilder: reference, name: "base"}
		i.Add("echo 1", "echo {{ .Version }}")
		cmds, e := i.Commands()
		So(e, ShouldBeNil)
		So(len(cmds), ShouldEqual, 2)

		So(i.Compile(), ShouldBeNil)

		cmds, e = i.Commands()
		So(len(cmds), ShouldEqual, 2)

		So(cmds[0].Shell(), ShouldEqual, "echo 1")
		So(cmds[1].Shell(), ShouldEqual, "echo 1.2")

		Convey("not being valid", func() {
			reference := &struct {
				Version string `urknall:"default=1.3"`
			}{}
			i := &task{taskBuilder: reference}
			i.Add("echo 1", "echo {{ .Version }}")
			cmds, e := i.Commands()
			So(e, ShouldBeNil)
			So(len(cmds), ShouldEqual, 2)

			So(i.Compile(), ShouldBeNil)

			cmds, e = i.Commands()
			So(len(cmds), ShouldEqual, 2)

			So(cmds[0].Shell(), ShouldEqual, "echo 1")
			So(cmds[1].Shell(), ShouldEqual, "echo 1.3")

		})
	})
}
