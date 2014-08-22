package urknall

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall/utils"
)

type packageImpl struct {
	tasks          []*task
	taskNames      map[string]struct{}
	reference      interface{} // used for rendering
	cacheKeyPrefix string
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
