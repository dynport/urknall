package urknall

import (
	"crypto/sha256"
	"fmt"

	"github.com/dynport/urknall/cmd"
)

func build(builder PackageBuilder) (Package, error) {
	p := &packageImpl{reference: builder}
	e := validatePackage(builder)
	if e != nil {
		return nil, e
	}
	builder.Build(p)
	return p, nil
}

func executeCommand(cmd cmd.Command, build *Build, checksumDir string) (e error) {
	sCmd := cmd.Shell()
	for _, env := range build.Env {
		sCmd = env + " " + sCmd
	}
	r := &remoteTaskRunner{build: build, cmd: sCmd, command: cmd, dir: checksumDir}
	return r.run()
}

func commandChecksum(c cmd.Command) string {
	if c, ok := c.(interface {
		Checksum() string
	}); ok {
		return c.Checksum()
	}
	s := sha256.New()
	s.Write([]byte(c.Shell()))
	return fmt.Sprintf("%x", s.Sum(nil))
}

func taskNameOfCommand(i interface{}) string {
	if c, ok := i.(interface {
		TaskName() string
	}); ok {
		return c.TaskName()
	}
	return ""
}
