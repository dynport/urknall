package cmd

import (
	"fmt"
	"os"
	"strings"
)

func InstallPackages(pkgs ...string) string {
	if len(pkgs) == 0 {
		panic("empty package list given")
	}
	return fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", strings.Join(pkgs, " "))
}

func And(cmds ...string) string {
	if len(cmds) == 0 {
		panic("empty list of commands given")
	}
	if len(cmds) == 1 {
		return cmds[0]
	}
	return fmt.Sprintf("{ %s; }", strings.Join(cmds, " && "))
}

func Or(cmds ...string) string {
	if len(cmds) == 0 {
		panic("empty list of commands given")
	}
	if len(cmds) == 1 {
		return cmds[0]
	}
	return fmt.Sprintf("{ %s; }", strings.Join(cmds, " || "))
}

func Mkdir(path, owner string, mode os.FileMode) string {
	if path == "" {
		panic("empty path given to mkdir")
	}

	cmds := []string{fmt.Sprintf("mkdir -p %s", path)}
	if owner != "" {
		cmds = append(cmds, fmt.Sprintf("chown %s %s", owner, path))
	}

	if mode != 0 {
		cmds = append(cmds, fmt.Sprintf("chmod %o %s", mode, path))
	}

	return And(cmds...)
}

func If(test, command string) string {
	if test == "" {
		panic("empty test given")
	}

	if command == "" {
		panic("empty command given")
	}

	return fmt.Sprintf("{ [[ %s ]] && %s; }", test, command)
}

func IfNot(test, command string) string {
	if test == "" {
		panic("empty test given")
	}

	if command == "" {
		panic("empty command given")
	}

	return fmt.Sprintf("{ [[ %s ]] || %s; }", test, command)
}
