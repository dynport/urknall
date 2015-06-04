package urknall

import (
	"testing"
)

type testPackageBuilder struct {
	Version string `urknall:"required=true"`
}

func (t *testPackageBuilder) Build(p Package) {
	ts := NewTask()
	ts.Add("echo base {{ .Version }}")
}

func TestBuildPackage(t *testing.T) {
	task := NewTask()

	if cmds, err := task.Add("apt-get update").Commands(); err != nil {
		t.Errorf("didn't expect and error, got %q", err)
	} else if len(cmds) != 1 {
		t.Errorf("expected task to have 1 comand, got %d", len(cmds))
	}

	pkg := &packageImpl{}
	pkg.AddTask("base", task)
	if len(pkg.tasks) != 1 {
		t.Errorf("expected pkg to have 1 task, got %d", len(pkg.tasks))
	}
}
