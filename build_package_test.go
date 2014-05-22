package urknall

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

type testPackageBuilder struct {
	Version string `urknall:"required=true"`
}

func (t *testPackageBuilder) Build(p Package) {
	p.AddTask(NewTask("base").Add("echo base {{ .Version }}"))
}

func TestBuildPackage(t *testing.T) {
	Convey("Build package", t, func() {
		task := NewTask("base").Add("apt-get update")
		So(task, ShouldNotBeNil)
		pkg := &packageImpl{}
		pkg.AddTask(task)

		So(len(pkg.Tasks()), ShouldEqual, 1)
		t.Log(task.Commands())
	})
}
