package cmd

import (
	"fmt"
	"github.com/dynport/zwo/host"
)

type DockerInitCommand struct {
	Command string // Command to be executed on container start.
}

func (diC *DockerInitCommand) Docker(host *host.Host) string {
	return fmt.Sprintf("CMD %s", diC.Command)
}

func (diC *DockerInitCommand) Shell(host *host.Host) string {
	return ""
}

func (diC *DockerInitCommand) Logging(host *host.Host) string {
	return fmt.Sprintf("[D.RUN  ] Adding docker init cmd: %.50s", diC.Command)
}
