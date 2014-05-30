package urknall

import "log"

func ExampleBuild_Run() {
	template := &ExampleTemplate{}
	target, e := NewLocalTarget()

	if e != nil {
		log.Fatal(e)
	}

	build := &Build{Target: target, Template: template}
	if e := build.Run(); e != nil {
		log.Fatal(e)
	}
}

// An example template function. This is helpful to render templates that don't
// need parameters like in the following ExampleTemplate.
func AnExampleTemplateFunc(pkg Package) {
	pkg.AddCommands("example", Shell("echo template func"))
}

// A simple template with some parameters.
type ExampleTemplate struct {
	Parameter string `urknall:"default=example"`
	Boolean   bool   `urknall:"required=true"`
}

// Templates must implement the Render function.
func (tmpl *ExampleTemplate) Render(pkg Package) {
	// Template parameters can be used in go's text/template style.
	pkg.AddCommands("base", Shell("echo {{ .Parameter }}"))
	if tmpl.Boolean { // Only add template function if Boolean value is true.
		pkg.AddTemplate("func", TemplateFunc(AnExampleTemplateFunc))
	}
}

// Need to implement a command. Those come with the default code created by the
// `urknall init` method, so in most cases this must not be done manually.
type ShellCmd struct {
	cmd string
}

func (c *ShellCmd) Logging() string {
	return c.cmd
}

func (c *ShellCmd) Shell() string {
	return c.cmd
}

// Helper function to easily create a ShellCmd.
func Shell(cmd string) Command {
	return &ShellCmd{cmd: cmd}
}
