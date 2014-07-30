package main

import (
	"fmt"
	"strings"
)

// Upgrade the package cache and update the installed packages (using apt).
func UpdatePackages() *ShellCommand {
	return And("apt-get update", "DEBIAN_FRONTEND=noninteractive apt-get upgrade -y")
}

// Update the package cache for a given repository only. Repo selection is done
// via the name of apt's configuration file taken from /etc/apt/sources.list.d.
// This is much faster if you just added a repo and want to install software as
// you need not update all other packages too (which most probably happened
// just recently during provisioning).
func UpdateSelectedRepoPackages(repoConfigPath string) *ShellCommand {
	return &ShellCommand{
		Command: fmt.Sprintf(
			`apt-get update -o Dir::Etc::sourcelist="sources.list.d/%s" -o Dir::Etc::sourceparts="-" -o APT::Get::List-Cleanup="0"`,
			repoConfigPath,
		),
	}
}

// Install the given packages using apt-get. At least one package must be given (pkgs can be left empty).
func InstallPackages(pkg string, pkgs ...string) *ShellCommand {
	return &ShellCommand{
		Command: fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s %s", pkg, strings.Join(pkgs, " ")),
	}
}

// PinPackage pins package via dpkg --set-selections
func PinPackage(name string) *ShellCommand {
	return Shell(fmt.Sprintf(`echo "%s hold" | dpkg --set-selections`, name))
}

// StartOrRestart starts or restarts a service configured with upstart
func StartOrRestart(service string) *ShellCommand {
	return Shell(fmt.Sprintf("if status %s | grep running; then restart %s ; else start %s; fi", service, service, service))
}
