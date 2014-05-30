package urknall

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall/pubsub"
	"github.com/dynport/urknall/utils"
)

// A template is used to modularize the urknall setting. Templates are rendered
// into a package and during rendering the tasks can be added. See the Package
// description for information on how to manage tasks.
type Template interface {
	Render(pkg Package)
}

// The package is an interface. Users won't implement it themselves though. It
// provides the methods used to add tasks to a package (AddTemplate and
// AddCommands will internally create tasks, so that in most cases users won't
// have to manage those by themselves).
//
// Nesting of templates provides a lot of flexibility as different
// configurations can be used depending on the greater context.
//
// The first argument of all three Add methods is a string. These strings are
// used as identifiers for the hashing mechanism. They must be unique over all
// tasks! For nested templates the identifiers are concatenated using ".".
type Package interface {
	AddTemplate(string, Template)   // Add another template, nested below the current one.
	AddCommands(string, ...Command) // Add a new task from the given commands.
	AddTask(string, Task)           // Add the given tasks to the package with the given name.

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
