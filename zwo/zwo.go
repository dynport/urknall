// zwo basic package with all the bells and whistles.
//
// This package contains the zwo core, i.e. the most essential stuff. Primary goals were an easy to user interface with
// good support for type errors. There are provisioners that execute actions from a runlist in order to provision the
// target system.
//
// Actions are the core building blocks that exist in the form of functions. These functions are required for the
// actions to be reusable (like a generic action to install certain packages), to be configurable with regards to the
// host to be provisioned (must commands be run with sudo for example), and errors be handled nicely. Actions come in
// different flavors like 'command actions' that will execute plain shell commands (I'm hesitant to write plain simple,
// because those expressions can get pretty advanced) or 'file actions' that will write files to the remote machines.
// Those different flavors are used to support docker and shell clients.
//
// A runlist is a container that actions are added, too. There are methods to add actions of the different flavors (to
// be type save and not intermix different kinds of actions). Users will not have to create runlists; they are generated
// interally. There is the 'Compiler' interface that must be implemented by entities (lets call them packages) used to
// structure commands. The compile method will be given the runlist to fill with commands (see the base package for
// example).
//
// The provisioner will be given a list packages (entities implementing the 'Compiler' interface), compiles a runlist
// for each and will run those on the host to provision, using the required mechanisms depending on the targeted host.
package zwo

import (
	"github.com/dynport/gologger"
)

var logger = gologger.NewFromEnv()

func init() {
	logger.Start()
}
