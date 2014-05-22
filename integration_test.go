package urknall

import (
	"sort"
	"testing"

	"github.com/dynport/urknall/utils"
	. "github.com/smartystreets/goconvey/convey"
)

type BuildHost struct {
}

func (b *BuildHost) BuildPackage(p Package) {
	p.Add("staging", &Staging{})
}

type Staging struct {
	RubyVersion string `urknall:"default=2.1.2"`
}

func (s *Staging) BuildPackage(p Package) {
	p.Add("ruby-{{ .RubyVersion }}", &Ruby{Version: s.RubyVersion})
	p.Add("es", &ElasticSearch{})
}

type Ruby struct {
	Version string
}

type ElasticSearch struct {
}

func (e *ElasticSearch) BuildPackage(p Package) {
	p.Add("install", []string{"apt-get install elasticsearch"})
	p.Add("ruby", &Ruby{})
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

func (r *Ruby) BuildPackage(p Package) {
	p.Add("install", []string{"apt-get update", "apt-get install ruby -v {{ .Version }}"})
	p.Add("config", []Command{&testCommand{cmd: "echo {{ .Version }}"}})
	p.Add("plain", []string{"using version {{ .Version }}"})
}

func rcover(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatal(r)
	}
}

func TestIntegration(t *testing.T) {
	Convey("Integration test", t, func() {
		defer rcover(t)
		pkg := &BuildHost{}
		p, e := build(pkg)
		So(e, ShouldBeNil)
		So(p, ShouldNotBeNil)

		names := []string{}

		tasks := map[string]Task{}
		for _, task := range p.Tasks() {
			tasks[task.CacheKey()] = task
			names = append(names, task.CacheKey())
		}

		So(len(names), ShouldEqual, 7)

		sort.Strings(names)
		So(names[0], ShouldEqual, "staging.es.install")
		So(names[1], ShouldEqual, "staging.es.ruby.config")
		So(names[2], ShouldEqual, "staging.es.ruby.install")
		So(names[3], ShouldEqual, "staging.es.ruby.plain")
		So(names[4], ShouldEqual, "staging.ruby-2.1.2.config")
		So(names[5], ShouldEqual, "staging.ruby-2.1.2.install")
		So(names[6], ShouldEqual, "staging.ruby-2.1.2.plain")

		task := tasks["staging.ruby-2.1.2.config"]
		commands, e := task.Commands()
		So(e, ShouldBeNil)
		So(len(commands), ShouldEqual, 1)
		So(commands[0].Shell(), ShouldEqual, "echo 2.1.2")

		task = tasks["staging.ruby-2.1.2.plain"]
		commands, e = task.Commands()
		So(e, ShouldBeNil)
		So(len(commands), ShouldEqual, 1)
		So(commands[0].Shell(), ShouldEqual, "using version 2.1.2")
	})
}
