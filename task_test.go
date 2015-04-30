package urknall

import (
	"testing"
)

type vers struct {
	Version string
}

func (v *vers) Render(Package) {
}

func TestTaskImpl(t *testing.T) {
	reference := &vers{"1.2"}
	i := &task{taskBuilder: reference, name: "base"}
	i.Add("echo 1", "echo {{ .Version }}")

	if cmds, err := i.Commands(); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if len(cmds) != 2 {
		t.Errorf("expected %d commands, got %q", 2, len(cmds))
	}

	if err := i.Compile(); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	if cmds, err := i.Commands(); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if len(cmds) != 2 {
		t.Errorf("expected %d commands, got %q", 2, len(cmds))
	} else if cmds[0].Shell() != "echo 1" {
		t.Errorf("expected command %d to be %q, got %q", 0, "echo 1", cmds[0])
	} else if cmds[1].Shell() != "echo 1.2" {
		t.Errorf("expected command %d to be %q, got %q", 1, "echo 1.2", cmds[1])
	}
}

func TestInvalidTaskImpl(t *testing.T) {
	reference := &struct {
		genericPkg
		Version string `urknall:"default=1.3"`
	}{}
	i := &task{taskBuilder: reference}
	i.Add("echo 1", "echo {{ .Version }}")

	if cmds, err := i.Commands(); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if len(cmds) != 2 {
		t.Errorf("expected %d commands, got %q", 2, len(cmds))
	}

	if err := i.Compile(); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	if cmds, err := i.Commands(); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if len(cmds) != 2 {
		t.Errorf("expected %d commands, got %q", 2, len(cmds))
	} else if cmds[0].Shell() != "echo 1" {
		t.Errorf("expected command %d to be %q, got %q", 0, "echo 1", cmds[0])
	} else if cmds[1].Shell() != "echo 1.3" {
		t.Errorf("expected command %d to be %q, got %q", 1, "echo 1.3", cmds[1])
	}
}
