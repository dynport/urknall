package main

import (
	"fmt"
	"time"
)

// Wait for the given path to appear. Break and fail if it doesn't appear after the given number of seconds.
func WaitForFile(path string, timeout time.Duration) *ShellCommand {
	t := 10 * timeout.Seconds()
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && [ ! -e %s ]; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"file %s did not appear\" 1>&2 && exit 1; }; }",
		int64(t), path, int64(t), path)
	return &ShellCommand{
		Command: cmd,
	}
}

// Wait for the given unix file socket to appear. Break and fail if it doesn't appear after the given number of seconds.
func WaitForUnixSocket(path string, timeout time.Duration) *ShellCommand {
	t := 10 * timeout.Seconds()
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && ! { netstat -lx | grep \"%s$\"; }; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"socket %s did not appear\" 1>&2 && exit 1; }; }",
		int64(t), path, int64(t), path)
	return &ShellCommand{
		Command: cmd,
	}
}
