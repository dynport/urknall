package urknall

import (
	"testing"
)

type customCommand struct {
	Content string
}

func (cc *customCommand) Shell() string {
	return "cc: " + cc.Content
}
func (cc *customCommand) Logging() string {
	return ""
}

type somePackage struct {
	SField string
	IField int
}

func (sp *somePackage) Render(Package) {
}

func TestAddCommand(t *testing.T) {
	rl := &task{taskBuilder: &somePackage{SField: "something", IField: 1}}

	rl.Add(`string with "{{ .SField }}" and "{{ .IField }}"`)

	c := rl.commands[len(rl.commands)-1].command
	if sc, ok := c.(*stringCommand); !ok {
		t.Errorf("expect ok, wasn't")
	} else if sc.cmd != `string with "something" and "1"` {
		t.Errorf("expect %q, got %q", `string with "something" and "1"`, sc.cmd)
	}
}

func TestAddStringCommand(t *testing.T) {
	rl := &task{taskBuilder: &somePackage{SField: "something", IField: 1}}

	baseCommand := stringCommand{cmd: `string with "{{ .SField }}" and "{{ .IField }}"`}

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("expected a panic, got none!")
			}
		}()
		rl.Add(baseCommand)

	}()

	rl.Add(&baseCommand)
	c := rl.commands[len(rl.commands)-1].command
	if sc, ok := c.(*stringCommand); !ok {
		t.Errorf("expect ok, wasn't")
	} else if sc.cmd != `string with "something" and "1"` {
		t.Errorf("expect %q, got %q", `string with "something" and "1"`, sc.cmd)
	}
}
