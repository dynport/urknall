package urknall

import (
	"fmt"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
	"runtime/debug"
)

// A runlist is a container for commands. Use the following methods to add new commands.
type Runlist struct {
	commands []cmd.Command
	pkg      Package
	name     string // Name of the compilable.
}

func (rl *Runlist) Add(c interface{}) {
	switch t := c.(type) {
	case *cmd.ShellCommand:
		t.Command = utils.MustRenderTemplate(t.Command, rl.pkg)
		rl.commands = append(rl.commands, t)
	case *cmd.FileCommand:
		t.Content = utils.MustRenderTemplate(t.Content, rl.pkg)
		rl.commands = append(rl.commands, t)
	case cmd.Command:
		rl.commands = append(rl.commands, t)
	case string:
		// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
		// explicitly renders the template.
		rl.Add(&cmd.ShellCommand{Command: t})
	default:
		panic(fmt.Sprintf("type %T not supported!", t))
	}
}

func (rl *Runlist) compile() (e error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to precompile package: %v", rl.name)
			}
			logger.Info(e.Error())
			logger.Debug(string(debug.Stack()))
		}
	}()

	if e = validatePackage(rl.pkg); e != nil {
		return e
	}

	rl.pkg.Package(rl)
	logger.Debugf("Precompiled package %s", rl.name)
	return nil
}
