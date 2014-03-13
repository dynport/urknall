package urknall

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/dynport/urknall/cmd"
)

// A runlist is a container for commands. Use the following methods to add new commands.
type Runlist struct {
	commands []cmd.Command

	name string  // Name of the compilable.
	pkg  Package // only used for rendering templates
	host *Host   // this is just for logging
}

// Add commands (can also be given as string) or packages (commands will be extracted and added accordingly) to the
// runlist.
func (rl *Runlist) Add(cmds ...interface{}) {
	for _, c := range cmds {
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			rl.addCommand(&stringCommand{cmd: t})
		case cmd.Command:
			rl.addCommand(t)
		case Package:
			rl.addPackage(t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
}

// Add the given package's commands to the runlist.
func (rl *Runlist) addPackage(p Package) {
	r := &Runlist{pkg: p, host: rl.host}
	e := validatePackage(p)
	if e != nil {
		panic(e.Error())
	}
	p.Package(r)
	rl.commands = append(rl.commands, r.commands...)
}

// Add the given command to the runlist.
func (rl *Runlist) addCommand(c cmd.Command) {
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

func (rl *Runlist) compile() (e error) {
	m := &Message{runlist: rl, host: rl.host, key: MessageRunlistsPrecompile}
	m.publish("started")
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to precompile package %q: %s", rl.name, r)
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
