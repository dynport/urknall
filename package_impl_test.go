package urknall

import (
	"testing"
)

func TestPackageChecksum(t *testing.T) {
	p := &packageImpl{}
	p.AddCommands("test1", Shell("echo 1"), Shell("echo 2"))
	p.AddCommands("test2", Shell("echo 2"), Shell("echo 2"), Shell("echo 3"))
	if len(p.tasks) != 2 {
		t.Errorf("tasks should be 2, was %d", len(p.tasks))
	}
	task := p.tasks[0]
	if len(task.commands) != 2 {
		t.Errorf("commands should be 2, was %d", len(task.commands))
	}
	if ex, v := p.tasks[0].commands[1].Checksum(), "9f8f29bb80830f069e821de502ec94200481550c208751d49bc7465815fff4f5"; ex != v {
		t.Errorf("expected cs to be %q, was %q", ex, v)
	}
	if ex, v := p.tasks[1].commands[0].Checksum(), "9f8f29bb80830f069e821de502ec94200481550c208751d49bc7465815fff4f5"; ex != v {
		t.Errorf("expected cs to be %q, was %q", ex, v)
	}
}

func TestPackageImplSingleArg(t *testing.T) {
	pkg := &packageImpl{}
	pkg.AddCommands("test", &testCommand{"this is a test"})
	if len(pkg.tasks) != 1 {
		t.Errorf("expected %d tasks, got %d", 1, len(pkg.tasks))
	}

	c, err := pkg.tasks[0].Commands()
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if c[0].Shell() != "this is a test" {
		t.Errorf("expected %q, got %q", "this is a test", c[0].Shell())
	}

	pkg.AddCommands("test2", &testCommand{"testcmd"})
	if len(pkg.tasks) != 2 {
		t.Errorf("expected %d tasks, got %d", 2, len(pkg.tasks))
	}

	c, err = pkg.tasks[1].Commands()
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if c[0].Shell() != "testcmd" {
		t.Errorf("expected %q, got %q", "testcmd", c[0].Shell())
	}
}

func TestPackageImplMultipleArgs(t *testing.T) {
	pkg := &packageImpl{}
	pkg.AddCommands("test", &testCommand{"echo hello"}, &testCommand{"echo world"})
	tasks := pkg.tasks
	if len(pkg.tasks) != 1 {
		t.Errorf("expected %d tasks, got %d", 1, len(pkg.tasks))
	}

	task := tasks[0]
	if task.name != "test" {
		t.Errorf("expected task name to be %q, got %q", "test", task.name)
	}

	c, err := task.Commands()
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if c[0].Shell() != "echo hello" {
		t.Errorf("expected %q, got %q", "echo hello", c[0].Shell())
	}
	if c[1].Shell() != "echo world" {
		t.Errorf("expected %q, got %q", "echo world", c[1].Shell())
	}

	pkg.AddCommands("test2", &testCommand{"echo cmd"})
	tasks = pkg.tasks
	if len(pkg.tasks) != 2 {
		t.Errorf("expected %d tasks, got %d", 2, len(pkg.tasks))
	}

	task = tasks[1]
	if task.name != "test2" {
		t.Errorf("expected task name to be %q, got %q", "test2", task.name)
	}
	c, err = task.Commands()
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if len(c) != 1 {
		t.Errorf("expected %d command, got %d", 1, len(pkg.tasks))
	}
	if c[0].Shell() != "echo cmd" {
		t.Errorf("expected %q, got %q", "echo cmd", c[0].Shell())
	}
}

type testPackage struct {
	Array []string `urknall:"required=true"`
}

func (tp *testPackage) Render(pkg Package) {
	for i := range tp.Array {
		pkg.AddCommands(tp.Array[i], Shell("echo "+tp.Array[i]))
	}
}

func TestTemplateWithStringSliceRequired(t *testing.T) {
	pkg := &packageImpl{}
	names := []string{"foo", "bar", "baz"}
	pkg.AddTemplate("test", &testPackage{Array: names})
	if len(pkg.tasks) != 3 {
		t.Fatalf("expected %d tasks, got %d", 3, len(pkg.tasks))
	}

	for i := range names {
		if pkg.tasks[i].name != "test."+names[i] {
			t.Errorf("task %d: expected task name %q, got %q", i, "test."+names[i], pkg.tasks[i].name)
		}
	}
}
