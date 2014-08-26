package urknall

import (
	"sort"
	"testing"

	"github.com/dynport/urknall/utils"
	. "github.com/smartystreets/goconvey/convey"
)

type BuildHost struct {
}

func (b *BuildHost) Render(p Package) {
	p.AddTemplate("staging", &Staging{})
}

type Staging struct {
	RubyVersion string `urknall:"default=2.1.2"`
}

func (s *Staging) Render(p Package) {
	p.AddTemplate("ruby-{{ .RubyVersion }}", &Ruby{Version: s.RubyVersion})
	p.AddTemplate("es", &ElasticSearch{})
}

type Ruby struct {
	Version string
}

type ElasticSearch struct {
}

func (e *ElasticSearch) Render(p Package) {
	p.AddCommands("install", &testCommand{cmd: "apt-get install elasticsearch"})
	p.AddTemplate("ruby", &Ruby{})
}

type testCommand struct {
	cmd string
}

func (c *testCommand) Shell() string {
	return c.cmd
}

func (c *testCommand) Logging() string {
	return c.cmd
}

func (c *testCommand) Render(i interface{}) {
	c.cmd = utils.MustRenderTemplate(c.cmd, i)
}

func (r *Ruby) Render(p Package) {
	t := NewTask()
	t.Add("apt-get update", "apt-get install ruby -v {{ .Version }}")
	p.AddTask("install", t)
	p.AddCommands("config", &testCommand{cmd: "echo {{ .Version }}"})
}

func rcover(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatal(r)
	}
}

func TestIntegration(t *testing.T) {
	Convey("Integration test", t, func() {
		bh := &BuildHost{}
		p, e := renderTemplate(bh)
		So(e, ShouldBeNil)
		So(p, ShouldNotBeNil)

		names := []string{}

		tasks := map[string]Task{}

		for _, task := range p.tasks {
			tasks[task.name] = task
			names = append(names, task.name)
		}

		t.Logf("%#v", names)

		So(len(names), ShouldEqual, 5)

		sort.Strings(names)
		So(names[0], ShouldEqual, "staging.es.install")
		So(names[1], ShouldEqual, "staging.es.ruby.config")
		So(names[2], ShouldEqual, "staging.es.ruby.install")
		So(names[3], ShouldEqual, "staging.ruby-2.1.2.config")
		So(names[4], ShouldEqual, "staging.ruby-2.1.2.install")

		task := tasks["staging.ruby-2.1.2.config"]
		commands, e := task.Commands()
		So(e, ShouldBeNil)
		So(len(commands), ShouldEqual, 1)
		So(commands[0].Shell(), ShouldEqual, "echo 2.1.2")
	})
}
