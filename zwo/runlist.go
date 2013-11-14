// The Runlist is a collection of associated commands used to provision a service.
//
// This is an internal representation of the commands. It is designed to make usage of the interface as easy as
// possible. It is in effect the result of the compilation call.
package zwo

import (
	"fmt"
	"github.com/dynport/zwo/host"
)

// A runlist is a container for commands. While those can have arbitrary intent, they should be closely related, for the
// sake of clarity and reusability.
type Runlist struct {
	actions []action
	host    *host.Host
	config  interface{}
	name    string // Name of the compilable.
}

// Add the given commands to the runlist.
//
// This function does all the error handling, i.e. calls the CommandF functions, thus retrieving the actual command or
// an error. If an error is seen that is propagated upwards.
func (rl *Runlist) AddCommands(cmds ...CommandActionFunc) (e error) {
	if len(cmds) == 0 {
		return fmt.Errorf("empty list of commands given")
	}
	for i := range cmds {
		c, e := cmds[i](rl.host, rl.config)
		if e != nil {
			return nil
		}
		rl.actions = append(rl.actions, c)
	}
	return nil
}

func (rl *Runlist) AddFiles(cmds ...FileActionFunc) (e error) {
	if len(cmds) == 0 {
		return fmt.Errorf("empty list of files given")
	}
	for i := range cmds {
		c, e := cmds[i](rl.host, rl.config)
		if e != nil {
			return nil
		}
		rl.actions = append(rl.actions, c)
	}
	return nil
}

// The configuration is used to expand the templates used for the commands, i.e. all fields and methods of the given
// entity are available in the template string (using the common "{{ .Something }}" notation, see text/template for more
// information).
func (rl *Runlist) setConfig(cfg interface{}) {
	rl.config = cfg
}

func (rl *Runlist) setName(name string) {
	rl.name = name
}

func (rl *Runlist) getName() (name string) {
	return rl.name
}
