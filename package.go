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
	AddCommands(string, ...command)
	AddTask(Task)

	Build(*Build) error
}

type packageImpl struct {
	tasks          []Task
	taskNames      map[string]struct{}
	reference      interface{} // used for rendering
	cacheKeyPrefix string
}

func (pkg *packageImpl) Build(build *Build) error {
	ct, e := build.buildChecksumTree()
	if e != nil {
		return e
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

func (pkg *packageImpl) AddCommands(name string, cmds ...command) {
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
	pkg.addTask(t, false)
}

func (pkg *packageImpl) AddTemplate(name string, tpl Template) {
	if pkg.cacheKeyPrefix != "" {
		name = pkg.cacheKeyPrefix + "." + name
	}
	name = utils.MustRenderTemplate(name, pkg.reference)
	e := validatePackage(tpl)
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
		pkg.addTask(task, false)
	}
}

func (pkg *packageImpl) addTask(task Task, addPrefix bool) {
	if addPrefix {
		name := task.Key()
		if pkg.cacheKeyPrefix != "" {
			name = pkg.cacheKeyPrefix + "." + task.Key()
		}
		name = utils.MustRenderTemplate(name, pkg.reference)
		task.SetKey(name)
	}
	pkg.validateTaskName(task.Key())
	pkg.taskNames[task.Key()] = struct{}{}
	pkg.tasks = append(pkg.tasks, task)
}

func (pkg *packageImpl) AddTask(task Task) {
	pkg.addTask(task, true)
}

func (pkg *packageImpl) precompile() (e error) {
	for _, task := range pkg.tasks {
		c, e := task.Commands()
		if e != nil {
			return e
		}
		if len(c) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", task.Key())
		}

		if tc, ok := task.(interface {
			Compile() error
		}); ok {
			e := tc.Compile()
			if e != nil {
				return e
			}
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
