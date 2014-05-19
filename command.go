package urknall

import (
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
)

type rawCommand struct {
	cmd.Command        // The command to be executed.
	checksum    string // The checksum of the command.
	task        *taskImpl
}

func (cmd *rawCommand) execute(build *Build, checksumDir string) (e error) {
	sCmd := cmd.Shell()
	for _, env := range build.Env {
		sCmd = env + " " + sCmd
	}
	r := &remoteTaskRunner{build: build, cmd: sCmd, rawCommand: cmd, dir: checksumDir}
	return r.run()
}

type stringCommand struct {
	cmd string
}

func (sc *stringCommand) Shell() string {
	return sc.cmd
}

func (sc *stringCommand) Logging() string {
	return "[COMMAND] " + sc.cmd
}

func (sc *stringCommand) Render(i interface{}) {
	sc.cmd = utils.MustRenderTemplate(sc.cmd, i)
}
