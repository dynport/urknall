// zwo provides everything necessary to provision machines, i.e. the  mechanisms required to run a set of commands
// somewhere (wherever that is, some bare metal or a docker container). These commands can be encapsulated in a package
// that has some configuration so that reuse is possible. There is even annotation based validation.
//
// Every package adds its raw commands into a runlist that is precompiled (to find errors prior to running the first
// remote command), has variable substitution for some commands (the package's fields can be used in commands rendered
// by go's templating mechanisms), and run on the respective host. This allows for provisioning packages in different
// configurations on different hosts.
//
// For each package a caching mechanism is used, so repeated provisioning of the same package will only run the commands
// necessary (a changed command and all subsequent ones). This allows for extension and modification of the
// underlying host and takes away the burden of writing idempotent commands. But in most cases it's more favorable to
// have throw away resources, that can easily replaced by a fresh one provisioned from ground up.
//
// Actions are the core building blocks that exist in the form of functions. These functions are required for the
// actions to be reusable (like a generic action to install certain packages), to be configurable with regards to the
// host to be provisioned (must commands be run with sudo for example), and errors be handled nicely. Actions come in
// different flavors like "command actions" that will execute plain shell commands (I'm hesitant to write plain simple,
// because those expressions can get pretty advanced) or "file actions" that will write files to the remote machines.
// Those different flavors are used to support docker and shell clients.
//
// A runlist is a container that actions are added, too. There are methods to add actions of the different flavors (to
// be type save and not intermix different kinds of actions). Users will not have to create runlists; they are generated
// interally. There is the "Compiler" interface that must be implemented by entities (lets call them packages) used to
// structure commands. The compile method will be given the runlist to fill with commands (see the base package for
// example).
//
// The provisioner will be given a list of packages (entities implementing the "Compiler" interface), compiles a runlist
// for each and will run those on the host to provision, using the required mechanisms depending on the targeted host.
package zwo

import (
	"github.com/dynport/gologger"
)

var logger = gologger.NewFromEnv()

func init() {
	logger.Start()
}
