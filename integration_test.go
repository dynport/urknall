package urknall

import (
	"sort"
	"testing"

	"github.com/dynport/urknall/utils"
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
	bh := &BuildHost{}
	p, e := renderTemplate(bh)
	if e != nil {
		t.Errorf("didn't expect an error")
	}
	if p == nil {
		t.Errorf("didn't expect the template to be nil")
	}

	names := []string{}

	tasks := map[string]Task{}

	for _, task := range p.tasks {
		tasks[task.name] = task
		names = append(names, task.name)
	}
	sort.Strings(names)

	if len(names) != 5 {
		t.Errorf("expected 5 names, got %d", len(names))
	}

	tt := []string{"staging.es.install", "staging.es.ruby.config", "staging.es.ruby.install", "staging.ruby-2.1.2.config", "staging.ruby-2.1.2.install"}
	for i := range tt {
		if names[i] != tt[i] {
			t.Errorf("expected names[%d] = %q, got %q", i, tt[i], names[i])
		}
	}

	task := tasks["staging.ruby-2.1.2.config"]
	commands, e := task.Commands()
	if e != nil {
		t.Errorf("didn't expect an error, got %s", e)
	}
	if len(commands) != 1 {
		t.Errorf("expected to find 1 command, got %d", len(commands))
	}
	if commands[0].Shell() != "echo 2.1.2" {
		t.Errorf("expected first command to be %q, got %q", "echo 2.1.2", commands[0].Shell())
	}
}
