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

	name string // Name of the compilable.
	task Tasker // only used for rendering templates
}

func (pkg *Task) Name() string {
	return pkg.name
}

func (p *Task) tasks() []*taskData {
	tasks := make([]*taskData, 0, len(p.commands))

	cmdHash := sha256.New()
	for i := range p.commands {
		rawCmd := p.commands[i].Shell()
		cmdHash.Write([]byte(rawCmd))

		task := &taskData{runlist: p, command: p.commands[i], checksum: fmt.Sprintf("%x", cmdHash.Sum(nil))}
		tasks = append(tasks, task)
	}
	return tasks
}

// Add commands (can also be given as string) or packages (commands will be extracted and added accordingly) to the
// runlist.
func (pkg *Task) Add(first interface{}, others ...interface{}) {
	all := append([]interface{}{first}, others...)
	for _, c := range all {
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			pkg.AddCommand(&stringCommand{cmd: t})
		case cmd.Command:
			pkg.AddCommand(t)
		case Tasker:
			pkg.AddPackage(t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
}

// Add the given package's commands to the runlist.
func (pkg *Task) AddPackage(p Tasker) {
	r := &Task{task: p}
	e := validatePackage(p)
	if e != nil {
		panic(e.Error())
	}
	p.Package(r)
	pkg.commands = append(pkg.commands, r.commands...)
}

// Add the given command to the runlist.
func (pkg *Task) AddCommand(c cmd.Command) {
	if pkg.task != nil {
		if renderer, ok := c.(cmd.Renderer); ok {
			renderer.Render(pkg.task)
		}
		if validator, ok := c.(cmd.Validator); ok {
			if e := validator.Validate(); e != nil {
				panic(e.Error())
			}
		}
	}
	pkg.commands = append(pkg.commands, c)
}

func (pkg *Task) compile() (e error) {
	m := &pubsub.Message{RunlistName: pkg.Name(), Key: pubsub.MessageRunlistsPrecompile}
	m.Publish("started")
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to precompile package: %v %q", pkg.name, r)
			}
			m.Error = e
			m.Stack = string(debug.Stack())
			m.Publish("panic")
			log.Printf("ERROR: %s", r)
			log.Print(string(debug.Stack()))
		}
	}()

	if e = validatePackage(pkg.task); e != nil {
		return e
	}
	pkg.task.Package(pkg)
	m.Publish("finished")
	return nil
}
