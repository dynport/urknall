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

// Add the given command actions to the runlist.
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

// Add the given files actions to the runlist.
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

// For the caching mechanism a unique identifier for each runlist is required. This identifier is set internally by the
// provisioner.
func (rl *Runlist) setName(name string) {
	rl.name = name
}

func (rl *Runlist) getName() (name string) {
	return rl.name
}
