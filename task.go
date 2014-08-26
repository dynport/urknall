package urknall

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

// A task is a list of commands. Each task is cached internally, i.e. if an
// command has been executed already, none of the preceding tasks has changed
// and neither the command itself, then it won't be executed again. This
// enhances performance and removes the burden of writing idempotent commands.
type Task interface {
	Add(cmds ...interface{}) Task
	Commands() ([]cmd.Command, error)
}

// Create a task. This is available to provide maximum flexibility, but
// shouldn't be required very often. The resulting task can be registered to an
// package using the AddTask method.
func NewTask() Task {
	return &task{}
}

type task struct {
	commands []*commandWrapper

	name        string   // Name of the compilable.
	taskBuilder Template // only used for rendering templates TODO(gf): rename

	compiled  bool
	validated bool
}

func (t *task) Commands() (cmds []cmd.Command, e error) {
	if e = t.Compile(); e != nil {
		return nil, e
	}

	for _, c := range t.commands {
		cmds = append(cmds, c.command)
	}

	return cmds, nil
}

func (task *task) Add(cmds ...interface{}) Task {
	for _, c := range cmds {
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			task.addCommand(&stringCommand{cmd: t})
		case cmd.Command:
			task.addCommand(t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
	return task
}

func (task *task) validate() error {
	if !task.validated {
		if task.taskBuilder == nil {
			return nil
		}
		e := validateTemplate(task.taskBuilder)
		if e != nil {
			return e
		}
		task.validated = true
	}
	return nil
}

// Add the given command to the runlist.
func (task *task) addCommand(c cmd.Command) {
	if task.taskBuilder != nil {
		e := task.validate()
		if e != nil {
			panic(e.Error())
		}
		if renderer, ok := c.(cmd.Renderer); ok {
			renderer.Render(task.taskBuilder)
		}
		if validator, ok := c.(cmd.Validator); ok {
			if e := validator.Validate(); e != nil {
				panic(e.Error())
			}
		}
	}
	task.commands = append(task.commands, &commandWrapper{command: c})
}

func (task *task) Compile() (e error) {
	if task.compiled {
		return nil
	}
	m := message(pubsub.MessageRunlistsPrecompile, "", task.name)
	m.Publish("started")
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to precompile package: %v %q", task.name, r)
			}
			m.Error = e
			m.Stack = string(debug.Stack())
			m.Publish("panic")
			log.Printf("ERROR: %s", r)
			log.Print(string(debug.Stack()))
		}
	}()

	e = task.validate()
	if e != nil {
		return e
	}
	m.Publish("finished")
	task.compiled = true
	return nil
}

type anonymousTask struct {
	cmds []interface{}
}

func (anon *anonymousTask) BuildTask(pkg Task) {
	for i := range anon.cmds {
		pkg.Add(anon.cmds[i])
	}
}
