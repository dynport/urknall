package zwo

import (
	"fmt"
	"github.com/dynport/zwo/host"
	"strings"
)

type commandAction struct {
	cmd  string
	user string
	host *host.Host
}

func (rCmd *commandAction) Docker() string {
	return fmt.Sprintf("RUN %s", rCmd.Plain())
}

func (rCmd *commandAction) Shell() string {
	cmdBuilder := 0

	if rCmd.user != "" && rCmd.user != "root" {
		cmdBuilder = 1
	}

	if rCmd.host.IsSudoRequired() {
		cmdBuilder += 2
	}

	switch cmdBuilder {
	case 0:
		return rCmd.cmd
	case 1:
		return fmt.Sprintf("su -l %s <<EOF\n%s\nEOF\n", rCmd.user, rCmd.cmd)
	case 2:
		return fmt.Sprintf("sudo bash <<EOF\n%s\nEOF\n", rCmd.cmd)
	case 3:
		return fmt.Sprintf("sudo -- su -l %s <<EOF\n%s\nEOF\n", rCmd.user, rCmd.cmd)
	}
	panic("should never be reached")
}

func (rCmd *commandAction) Plain() string {
	return rCmd.cmd
}

func (rCmd *commandAction) Logging() string {
	s := []string{"[COMMAND]"}

	if rCmd.host.IsSudoRequired() {
		s = append(s, "[SUDO]")
	}

	if rCmd.user != "" && rCmd.user != "root" {
		s = append(s, fmt.Sprintf("[SU:%s]", rCmd.user))
	}

	s = append(s, fmt.Sprintf("# %s", rCmd.cmd))

	return strings.Join(s, " ")
}
