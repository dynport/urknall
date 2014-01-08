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

	name string  // Name of the compilable.
	pkg  Package // only used for rendering templates
	host *Host   // this is just for logging
}

func (runlist *Runlist) Name() string {
	return runlist.name
}

func (rl *Runlist) Add(first interface{}, others ...interface{}) {
	all := append([]interface{}{first}, others...)
	for _, c := range all {
		switch t := c.(type) {
		case string:
			// No explicit expansion required as the function is called recursively with a ShellCommand type, that has
			// explicitly renders the template.
			rl.AddCommand(&cmd.ShellCommand{Command: t})
		case cmd.Command:
			rl.AddCommand(t)
		case Package:
			rl.AddPackage(t)
		default:
			panic(fmt.Sprintf("type %T not supported!", t))
		}
	}
}

func (rl *Runlist) AddPackage(p Package) {
	r := &Runlist{pkg: p, host: rl.host}
	e := validatePackage(p)
	if e != nil {
		panic(e.Error())
	}
	p.Package(r)
	rl.commands = append(rl.commands, r.commands...)
}

func (rl *Runlist) AddCommand(c cmd.Command) {
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

func (rl *Runlist) compileWithBinaryPackages() (e error) {
	return rl.compile(true)
}

func (rl *Runlist) compileWithoutBinaryPackages() (e error) {
	return rl.compile(false)
}

func (rl *Runlist) compile(useBinaryPkg bool) (e error) {
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

	if _, ok := rl.pkg.(BinaryPackage); ok && useBinaryPkg {
		if rl.host.BinaryPackageRepository != "" {
			m.publish(fmt.Sprintf("going to use a binary package from %q", rl.host.BinaryPackageRepository))
			rl.installBinaryPackage()
			m.publish("finished")
			return nil
		}
		m.publish(fmt.Sprintf("package %q's binary not used, as no repository defined. Going to build.", rl.name))
	}
	if e = validatePackage(rl.pkg); e != nil {
		return e
	}
	rl.pkg.Package(rl)
	m.publish("finished")
	return nil
}
