package urknall

import (
	"testing"
)

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
