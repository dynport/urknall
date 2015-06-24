package target

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Create a target for local provisioning.
func NewLocalTarget() *localTarget {
	return &localTarget{}
}

type localTarget struct {
	cachedUser string
}

func (c *localTarget) String() string {
	return "LOCAL"
}

func (c *localTarget) User() string {
	if c.cachedUser == "" {
		var err error
		c.cachedUser, err = whoami()
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	return c.cachedUser
}

func whoami() (string, error) {
	u := os.Getenv("$USER")
	if u != "" {
		return u, nil
	}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}
	cmd := exec.Command("whoami")
	cmd.Stdout = out
	cmd.Stderr = err
	e := cmd.Run()
	if e != nil {
		return "", fmt.Errorf("error reading login name: err=%q out=%q e=%q", err.String(), out.String(), e)
	}
	return strings.TrimSpace(out.String()), nil
}

func (c *localTarget) Command(cmd string) (ExecCommand, error) {
	return &localCommand{
		command: exec.Command("bash", "-c", cmd),
	}, nil
}

func (c *localTarget) Reset() (e error) {
	return nil
}

type localCommand struct {
	command *exec.Cmd
}

func (c *localCommand) StdoutPipe() (io.Reader, error) {
	return c.command.StdoutPipe()
}

func (c *localCommand) StderrPipe() (io.Reader, error) {
	return c.command.StderrPipe()
}

func (c *localCommand) StdinPipe() (io.WriteCloser, error) {
	return c.command.StdinPipe()
}

func (c *localCommand) SetStdout(w io.Writer) {
	c.command.Stdout = w
}

func (c *localCommand) SetStderr(w io.Writer) {
	c.command.Stderr = w
}

func (c *localCommand) SetStdin(r io.Reader) {
	c.command.Stdin = r
}

func (c *localCommand) Wait() error {
	return c.command.Wait()
}

func (c *localCommand) Start() error {
	return c.command.Start()
}

func (c *localCommand) Run() error {
	return c.command.Run()
}
