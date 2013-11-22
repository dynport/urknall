package cmd

import (
	"fmt"
	"os"
	"path"
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

func download(url string) string {
	cmd := And(
		"mkdir -p /tmp/downloads",
		"cd /tmp/downloads",
		fmt.Sprintf("curl -SsfLO %s", url))
	return cmd
}

func DownloadToFile(url, destination, owner string, mode os.FileMode) string {
	cmds := []string{}
	cmds = append(cmds, download(url))
	cmds = append(cmds, fmt.Sprintf("mv /tmp/downloads/%s %s", filenameFromUrl(url), destination))
	if owner != "" || owner != "root" {
		cmds = append(cmds, Or(
			If(fmt.Sprintf("-f %s", destination), fmt.Sprintf("chown %s %s", owner, destination)),
			If(fmt.Sprintf("-d %s", destination), fmt.Sprintf("chown %s %s/%s", owner, destination, filenameFromUrl(url))),
			And("echo \"Couldn't determine target\"", "exit 1")))
	}
	if mode != 0 {
		cmds = append(cmds, Or(
			If(fmt.Sprintf("-f %s", destination), fmt.Sprintf("chmod %o %s", mode, destination)),
			If(fmt.Sprintf("-d %s", destination), fmt.Sprintf("chmod %o %s/%s", mode, destination, filenameFromUrl(url))),
			And("echo \"Couldn't determine target\"", "exit 1")))
	}
	return And(cmds...)
}

func DownloadAndExtract(url, targetDir string) string {
	return And(
		download(url),
		If(fmt.Sprintf("! -d %s", targetDir), Mkdir(targetDir, "", 0)),
		fmt.Sprintf("cd %s", targetDir),
		fmt.Sprintf("tar xvfz /tmp/downloads/%s", filenameFromUrl(url)))
}

func filenameFromUrl(url string) string {
	return path.Base(url)
}
