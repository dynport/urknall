package urknall

import (
	"fmt"

	"github.com/dynport/urknall/cmd"
)

type rawCommand struct {
	cmd.Command        // The command to be executed.
	checksum    string // The checksum of the command.
	task        *Task
}

func (cmd *rawCommand) execute(build *Build, checksumDir string) (e error) {
	sCmd := fmt.Sprintf("sh -x -e -c %q", cmd.Shell())
	for _, env := range build.Env {
		sCmd = env + " " + sCmd
	}
	r := &remoteTaskRunner{build: build, cmd: sCmd, rawCommand: cmd, dir: checksumDir}
	return r.run()
}
