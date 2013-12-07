package cmd

import (
	"github.com/dynport/zwo/host"
)

// The Commander interface must be implemented by the different command types. It is used to run the command in different
// contexts, i.e. running a command in the shell is different from docker or logging.
type Commander interface {
	Docker(host *host.Host) string  // Used for executing the action in a docker context.
	Shell(host *host.Host) string   // Used for executing the action in a shell (locally or via ssh).
	Logging(host *host.Host) string // Get string used for logging.
}
