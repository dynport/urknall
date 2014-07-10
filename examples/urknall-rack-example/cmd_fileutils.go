package main

import (
	"fmt"
	"os"
)

// Create the given directory with the owner and file permissions set accordingly. If the last two options are set to
// go's default values nothing is done.
func Mkdir(path, owner string, permissions os.FileMode) *ShellCommand {
	if path == "" {
		panic("empty path given to mkdir")
	}

	mkdirCmd := fmt.Sprintf("mkdir -p %s", path)

	optsCmds := make([]interface{}, 0, 2)
	if owner != "" {
		optsCmds = append(optsCmds, fmt.Sprintf("chown %s %s", owner, path))
	}

	if permissions != 0 {
		optsCmds = append(optsCmds, fmt.Sprintf("chmod %o %s", permissions, path))
	}

	return And(mkdirCmd, optsCmds...)
}
