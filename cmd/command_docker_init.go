package cmd

import (
	"fmt"
)

// A command to be executed when a container is started. This is equivalent to the "UpstartCommand" of the bare metal
// (or virtual machine) provisioning.
//
// TODO: Generalize the "DockerInitCommand" so that it can set all the commands supported by dockerfiles.
type DockerInitCommand struct {
	Command string // Command to be executed on container start.
}

func (diC *DockerInitCommand) Docker() string {
	return fmt.Sprintf("CMD %s", diC.Command)
}

func (diC *DockerInitCommand) Shell() string {
	return ""
}

func (diC *DockerInitCommand) Logging() string {
	return fmt.Sprintf("[D.RUN  ] Adding docker init cmd: %.50s", diC.Command)
}
