package main

import (
	"fmt"
	"strings"
)

// Upgrade the package cache and update the installed packages (using apt).
func UpdatePackages() *ShellCommand {
	return And("apt-get update", "DEBIAN_FRONTEND=noninteractive apt-get upgrade -y")
}

// Install the given packages using apt-get. At least one package must be given (pkgs can be left empty).
func InstallPackages(pkg string, pkgs ...string) *ShellCommand {
	return &ShellCommand{
		Command: fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s %s", pkg, strings.Join(pkgs, " ")),
	}
}
