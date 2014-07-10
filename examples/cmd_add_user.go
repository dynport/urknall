package main

import "fmt"

// AddUser adds a new linux user (normal or system user) if it does not exist already
func AddUser(name string, systemUser bool) *ShellCommand {
	testForUser := "id " + name + " 2>&1 > /dev/null"
	userAddOpts := ""
	if systemUser {
		userAddOpts = "--system"
	} else {
		userAddOpts = "-m -s /bin/bash"
	}
	return Or(testForUser, fmt.Sprintf("useradd %s %s", userAddOpts, name))
}
