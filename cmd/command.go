package cmd

import (
	"github.com/dynport/zwo/host"
)

// The Commander interface must be implemented by the different command types. It is used to run the command in
// different contexts, i.e. either shell, docker, or logging. The last one comes in handy if a command's underlying as
// some command's doings are rather lengthy or cryptic, but the intent is described easily (like writing assets or
// file).
//
// Each of the interface's functions gets access to the configuration of the host, the command is executed on. This is
// for example required to determine whether sudo or su are required.
type Commander interface {
	Docker(host *host.Host) string  // Used for executing the action in a docker context.
	Shell(host *host.Host) string   // Used for executing the action in a shell (locally or via ssh).
	Logging(host *host.Host) string // Get string used for logging.
}
