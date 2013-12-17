package urknall

import (
	"fmt"
	"github.com/dynport/urknall/cmd"
	"log"
	"runtime/debug"
)

// A runlist is a container for commands. Use the following methods to add new commands.
type Runlist struct {
	commands []cmd.Command
	pkg      Package
	name     string // Name of the compilable.
}

func (runlist *Runlist) Name() string {
	return runlist.name
}

func (rl *Runlist) Add(first interface{}, others ...interface{}) {
	all := append([]interface{}{first}, others...)
	for _, c := range all {
		if renderer, ok := c.(cmd.Renderer); ok {
			renderer.Render(rl.pkg)
		}
		if validator, ok := c.(cmd.Validator); ok {
			if e := validator.Validate(); e != nil {
				panic(e.Error())
			}
		}
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			rl.Add(&cmd.ShellCommand{Command: t})
		case cmd.Command:
			rl.commands = append(rl.commands, t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
}

func (rl *Runlist) compile(host *Host) (e error) {
	m := &Message{runlist: rl, host: host, key: MessageRunlistsPrecompile}
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
