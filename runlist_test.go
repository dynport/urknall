package urknall

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Given a runlist for a certain package", t, func() {
		rl := &task{taskBuilder: &somePackage{SField: "something", IField: 1}}

		Convey("When a string is added", func() {
			rl.Add(`string with "{{ .SField }}" and "{{ .IField }}"`)
			Convey("Then the string is turned to a command and appended to the list of commands", func() {
				c := rl.commands[len(rl.commands)-1].command
				sc, ok := c.(*stringCommand)

				Convey("And the command is a string command", func() {
					So(ok, ShouldBeTrue)
				})
				Convey("And the command template itself was expanded", func() {
					So(sc.cmd, ShouldEqual, `string with "something" and "1"`)
				})
			})
		})

		Convey("Given a string command", func() {
			baseCommand := stringCommand{cmd: `string with "{{ .SField }}" and "{{ .IField }}"`}

			SkipConvey("When it is added to the runlist by value", func() {
				f := func() { rl.Add(baseCommand) }

				Convey("Then Add will panic", func() {
					So(f, ShouldPanic)
				})
			})

			Convey("When it is added by reference", func() {
				rl.Add(&baseCommand)
				c := rl.commands[len(rl.commands)-1].command
				sc, ok := c.(*stringCommand)

				Convey("Then the command is a string command", func() {
					So(ok, ShouldBeTrue)
				})
				Convey("And the command template itself was expanded", func() {
					So(sc.cmd, ShouldEqual, `string with "something" and "1"`)
				})
			})
		})
	})
}
