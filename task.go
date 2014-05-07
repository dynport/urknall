package urknall

import (
	"crypto/sha256"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

// A runlist is a container for commands. Use the following methods to add new commands.
type Task struct {
	commands []cmd.Command

	name string       // Name of the compilable.
	task TaskPackager // only used for rendering templates TODO(gf): rename
}

type TaskPackager interface {
	Package(*Task)
}

// Create a task from a set of commands without configuration.
func NewTask(cmds ...interface{}) *Task {
	return &Task{task: &anonymousTask{cmds: cmds}}
}

func (task *Task) rawCommands() []*rawCommand {
	rawCommands := make([]*rawCommand, 0, len(task.commands))

	cmdHash := sha256.New()
	for i := range task.commands {
		cmdHash.Write([]byte(task.commands[i].Shell()))
		rawCmd := &rawCommand{task: task, Command: task.commands[i], checksum: fmt.Sprintf("%x", cmdHash.Sum(nil))}
		rawCommands = append(rawCommands, rawCmd)
	}
	return rawCommands
}

func (task *Task) Add(cmds ...interface{}) {
	for _, c := range cmds {
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			task.addCommand(&stringCommand{cmd: t})
		case cmd.Command:
			task.addCommand(t)
		case TaskPackager:
			task.addPackage(t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
}

// Add the given package's commands to the runlist.
func (task *Task) addPackage(p TaskPackager) {
	r := &Task{task: p}
	e := validatePackage(p)
	if e != nil {
		panic(e.Error())
	}
	p.Package(r)
	task.commands = append(task.commands, r.commands...)
}

// Add the given command to the runlist.
func (task *Task) addCommand(c cmd.Command) {
	if task.task != nil {
		if renderer, ok := c.(cmd.Renderer); ok {
			renderer.Render(task.task)
		}
		if validator, ok := c.(cmd.Validator); ok {
			if e := validator.Validate(); e != nil {
				panic(e.Error())
			}
		}
	}
	task.commands = append(task.commands, c)
}

func (task *Task) compile() (e error) {
	m := &pubsub.Message{RunlistName: task.name, Key: pubsub.MessageRunlistsPrecompile}
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

	if e = validatePackage(task.task); e != nil {
		return e
	}
	task.task.Package(task) // TODO(gf): ouch
	m.Publish("finished")
	return nil
}

type anonymousTask struct {
	cmds []interface{}
}

func (anon *anonymousTask) Package(pkg *Task) {
	for i := range anon.cmds {
		pkg.Add(anon.cmds[i])
	}
}
