package urknall

import (
	"crypto/sha256"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/dynport/urknall/cmd"
)

// A runlist is a container for commands. Use the following methods to add new commands.
type Package struct {
	commands []cmd.Command

	name string   // Name of the compilable.
	pkg  Packager // only used for rendering templates
	host *Host    // this is just for logging
}

func (rl *Package) Name() string {
	return rl.name
}

func (p *Package) tasks() []*taskData {
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
func (rl *Package) Add(first interface{}, others ...interface{}) {
	all := append([]interface{}{first}, others...)
	for _, c := range all {
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			rl.AddCommand(&cmd.ShellCommand{Command: t})
		case cmd.Command:
			rl.AddCommand(t)
		case Packager:
			rl.AddPackage(t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
}

// Add the given package's commands to the runlist.
func (rl *Package) AddPackage(p Packager) {
	r := &Package{pkg: p, host: rl.host}
	e := validatePackage(p)
	if e != nil {
		panic(e.Error())
	}
	p.Package(r)
	rl.commands = append(rl.commands, r.commands...)
}

// Add the given command to the runlist.
func (rl *Package) AddCommand(c cmd.Command) {
	if rl.pkg != nil {
		if renderer, ok := c.(cmd.Renderer); ok {
			renderer.Render(rl.pkg)
		}
		if validator, ok := c.(cmd.Validator); ok {
			if e := validator.Validate(); e != nil {
				panic(e.Error())
			}
		}
	}
	rl.commands = append(rl.commands, c)
}

func (rl *Package) compile() (e error) {
	m := &Message{runlist: rl, host: rl.host, key: MessageRunlistsPrecompile}
	m.publish("started")
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to precompile package: %v %q", rl.name, r)
			}
			m.error_ = e
			m.stack = string(debug.Stack())
			m.publish("panic")
			log.Printf("ERROR: %s", r)
			log.Print(string(debug.Stack()))
		}
	}()

	if e = validatePackage(rl.pkg); e != nil {
		return e
	}
	rl.pkg.Package(rl)
	m.publish("finished")
	return nil
}
