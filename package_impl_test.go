package urknall

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPackageImpl(t *testing.T) {
	Convey("Test packageImpl", t, func() {
		Convey("add single arguments", func() {
			pkg := &packageImpl{}
			pkg.Add("test", "this is a test")
			So(len(pkg.Tasks()), ShouldEqual, 1)

			c, e := pkg.Tasks()[0].Commands()
			So(e, ShouldBeNil)
			So(c[0].Shell(), ShouldEqual, "this is a test")

			pkg.Add("test2", &testCommand{"testcmd"})
			So(len(pkg.Tasks()), ShouldEqual, 2)

			c, e = pkg.Tasks()[1].Commands()
			So(e, ShouldBeNil)
			So(c[0].Shell(), ShouldEqual, "testcmd")
		})

		Convey("add multiple arguments", func() {
			pkg := &packageImpl{}
			pkg.Add("test", "echo hello", "echo world")
			tasks := pkg.Tasks()
			So(len(tasks), ShouldEqual, 1)
			task := tasks[0]
			So(task.CacheKey(), ShouldEqual, "test")
			c, e := task.Commands()
			So(e, ShouldBeNil)
			So(c[0].Shell(), ShouldEqual, "echo hello")
			So(c[1].Shell(), ShouldEqual, "echo world")

			pkg.Add("test2", "echo hello", &testCommand{"echo cmd"})
			tasks = pkg.Tasks()
			So(len(tasks), ShouldEqual, 2)

			task = tasks[1]
			So(task.CacheKey(), ShouldEqual, "test2")
			c, e = task.Commands()
			So(e, ShouldBeNil)
			So(len(c), ShouldEqual, 2)
			So(c[1].Shell(), ShouldEqual, "echo cmd")
		})
	})
}
