package urknall

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall/pubsub"
	"github.com/dynport/urknall/utils"
)

type Template interface {
	Render(pkg Package)
}

type Package interface {
	AddTemplate(string, Template)
	AddCommands(string, ...Command)
	AddTask(string, Task)

	Build(*Build) error
}

type packageImpl struct {
	tasks          []*task
	taskNames      map[string]struct{}
	reference      interface{} // used for rendering
	cacheKeyPrefix string
}

func (pkg *packageImpl) Build(build *Build) error {
	e := build.prepare()
	if e != nil {
		return e
	}
	ct, e := build.buildChecksumTree()
	if e != nil {
		return fmt.Errorf("error building checksum tree: %s", e.Error())
	}

	for _, task := range pkg.tasks {
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvision, Hostname: build.hostname()}
		m.Publish("started")
		if e = build.buildTask(task, ct); e != nil {
			m.PublishError(e)
			return e
		}
		m.Publish("finished")
	}
	return nil
}

func (pkg *packageImpl) AddCommands(name string, cmds ...Command) {
	if pkg.cacheKeyPrefix != "" {
		name = pkg.cacheKeyPrefix + "." + name
	}
	name = utils.MustRenderTemplate(name, pkg.reference)
	t := &task{name: name}
	for _, c := range cmds {
		if r, ok := c.(renderer); ok {
			r.Render(pkg.reference)
		}
		t.Add(c)
	}
	pkg.addTask(t)
}

func (pkg *packageImpl) AddTemplate(name string, tpl Template) {
	if pkg.cacheKeyPrefix != "" {
		name = pkg.cacheKeyPrefix + "." + name
	}
	name = utils.MustRenderTemplate(name, pkg.reference)
	e := validateTemplate(tpl)
	if e != nil {
		panic(e)
	}
	if pkg.reference != nil {
		name = utils.MustRenderTemplate(name, pkg.reference)
	}
	pkg.validateTaskName(name)
	child := &packageImpl{cacheKeyPrefix: name, reference: tpl}
	tpl.Render(child)
	for _, task := range child.tasks {
		pkg.addTask(task)
	}
}

func (pkg *packageImpl) AddTask(name string, tsk Task) {
	if pkg.cacheKeyPrefix != "" {
		name = pkg.cacheKeyPrefix + "." + name
	}
	name = utils.MustRenderTemplate(name, pkg.reference)
	t := &task{name: name}
	cmds, e := tsk.Commands()
	if e != nil {
		panic(e)
	}
	for _, c := range cmds {
		t.Add(c)
	}
	pkg.addTask(t)
}

func (pkg *packageImpl) addTask(task *task) {
	pkg.validateTaskName(task.name)
	pkg.taskNames[task.name] = struct{}{}
	pkg.tasks = append(pkg.tasks, task)
}

func (pkg *packageImpl) precompile() (e error) {
	for _, task := range pkg.tasks {
		c, e := task.Commands()
		if e != nil {
			return e
		}
		if len(c) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", task.name)
		}

		e = task.Compile()
		if e != nil {
			return e
		}
	}

	return nil
}

func (pkg *packageImpl) validateTaskName(name string) {
	if name == "" {
		panic("package names must not be empty!")
	}

	if strings.Contains(name, " ") {
		panic(fmt.Sprintf(`package names must not contain spaces (%q does)`, name))
	}

	if pkg.taskNames == nil {
		pkg.taskNames = map[string]struct{}{}
	}

	if _, ok := pkg.taskNames[name]; ok {
		panic(fmt.Sprintf("package with name %q exists already", name))
	}
}
