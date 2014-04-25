package local

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/dynport/urknall"
)

type Host struct {
	cachedUser string
}

func (c *Host) String() string {
	return "LOCAL"
}

func (c *Host) User() string {
	if c.cachedUser == "" {
		out := &bytes.Buffer{}
		err := &bytes.Buffer{}
		cmd := exec.Command("whoami")
		cmd.Stdout = out
		cmd.Stderr = err
		e := cmd.Run()
		if e != nil {
			fmt.Printf("error reading login name: err=%q out=%q e=%q", err.String(), out.String, e)
			os.Exit(1)
		}
		c.cachedUser = out.String()
	}
	return c.cachedUser
}

func (c *Host) Command(cmd string) (urknall.Command, error) {
	return &Command{
		command: exec.Command("bash", "-c", cmd),
	}, nil
}
