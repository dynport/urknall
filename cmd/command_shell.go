package cmd

import (
	"fmt"
	"github.com/dynport/zwo/host"
	"os"
	"path"
	"strings"
)

type ShellCommand struct {
	Command string // Command to be executed in the shell.
	user    string // User to run the command as.
}

// Convenience function to run a command as a certain user. Please note that nested calls will not work (function will panic).
func AsUser(user string, i interface{}) *ShellCommand {
	switch c := i.(type) {
	case *ShellCommand:
		if c.isExecutedAsUser() {
			panic(`nesting "AsUser" calls not supported`)
		}
		c.user = user
		return c
	case string:
		return &ShellCommand{Command: c, user: user}
	default:
		panic(fmt.Sprintf(`type "%T" not supported`, c))
	}
}

// Upgrade the package cache and update the installed packages.
func UpdatePackages() *ShellCommand {
	return And("apt-get update", "DEBIAN_FRONTEND=noninteractive apt-get upgrade -y")
}

// Install the given packages.
func InstallPackages(pkgs ...string) *ShellCommand {
	if len(pkgs) == 0 {
		panic("empty package list given")
	}
	return &ShellCommand{
		Command: fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", strings.Join(pkgs, " ")),
	}
}

// Combine the given commands with "and" (all must succeed).
func And(cmds ...interface{}) *ShellCommand {
	if len(cmds) == 0 {
		panic("empty list of commands given")
	}

	cs := mergeSubCommands(cmds...)

	finalCommand := fmt.Sprintf("{ %s; }", strings.Join(cs, " && "))
	if len(cs) == 1 {
		finalCommand = cs[0]
	}
	return &ShellCommand{Command: finalCommand}
}

// Combine the given commands with "or" (try one after one, till the first works).
func Or(cmds ...interface{}) *ShellCommand {
	if len(cmds) == 0 {
		panic("empty list of commands given")
	}

	cs := mergeSubCommands(cmds...)

	finalCommand := fmt.Sprintf("{ %s; }", strings.Join(cs, " || "))
	if len(cs) == 1 {
		finalCommand = cs[0]
	}
	return &ShellCommand{Command: finalCommand}
}

func mergeSubCommands(cmds ...interface{}) (cs []string) {
	for i := range cmds {
		switch cmd := cmds[i].(type) {
		case *ShellCommand:
			if cmd.user != "" && cmd.user != "root" {
				panic("AsUser not supported in nested commands")
			}
			cs = append(cs, cmd.Command)
		case string:
			cs = append(cs, cmd)
		default:
			panic(fmt.Sprintf(`type "%T" not supported`, cmd))
		}
	}
	return cs
}

// Create the given directory with the owner and file permissions set accordingly.
func Mkdir(path, owner string, permissions os.FileMode) *ShellCommand {
	if path == "" {
		panic("empty path given to mkdir")
	}

	cmds := make([]interface{}, 0, 3)
	cmds = append(cmds, fmt.Sprintf("mkdir -p %s", path))
	if owner != "" {
		cmds = append(cmds, fmt.Sprintf("chown %s %s", owner, path))
	}

	if permissions != 0 {
		cmds = append(cmds, fmt.Sprintf("chmod %o %s", permissions, path))
	}

	return And(cmds...)
}

// If the tests succeeds run the given command (see "man test" for test syntax).
func If(test string, i interface{}) *ShellCommand {
	if test == "" {
		panic("empty test given")
	}

	baseCommand := "{ [[ %s ]] && %s; }"

	switch cmd := i.(type) {
	case *ShellCommand:
		cmd.Command = fmt.Sprintf(baseCommand, test, cmd.Command)
		return cmd
	case string:
		if cmd == "" {
			panic("empty command given")
		}
		return &ShellCommand{Command: fmt.Sprintf(baseCommand, test, cmd)}
	default:
		panic(fmt.Sprintf(`type "%T" not supported`, cmd))
	}
}

// If the tests does not succeed run the given command (see "man test" for test syntax).
func IfNot(test string, i interface{}) *ShellCommand {
	if test == "" {
		panic("empty test given")
	}

	baseCommand := "{ [[ %s ]] || %s; }"

	switch cmd := i.(type) {
	case *ShellCommand:
		cmd.Command = fmt.Sprintf(baseCommand, test, cmd.Command)
		return cmd
	case string:
		if cmd == "" {
			panic("empty command given")
		}
		return &ShellCommand{Command: fmt.Sprintf(baseCommand, test, cmd)}
	default:
		panic(fmt.Sprintf(`type "%T" not supported`, cmd))
	}
}

func download(url string) *ShellCommand {
	if url == "" {
		panic("empty url given")
	}
	return And(
		"mkdir -p /tmp/downloads",
		"cd /tmp/downloads",
		fmt.Sprintf("curl -SsfLO %s", url))
}

// Downlowad the URL to the destination with owner and permissions set accordingly.
func DownloadToFile(url, destination, owner string, permissions os.FileMode) *ShellCommand {
	cmds := make([]interface{}, 0, 4)
	cmds = append(cmds, download(url))

	if destination == "" {
		panic("empty destination given")
	}
	cmds = append(cmds, fmt.Sprintf("mv /tmp/downloads/%s %s", filenameFromUrl(url), destination))

	if owner != "" && owner != "root" {
		cmds = append(cmds, Or(
			If(fmt.Sprintf("-f %s", destination), fmt.Sprintf("chown %s %s", owner, destination)),
			If(fmt.Sprintf("-d %s", destination), fmt.Sprintf("chown %s %s/%s", owner, destination, filenameFromUrl(url))),
			And("echo \"Couldn't determine target\"", "exit 1")))
	}
	if permissions != 0 {
		cmds = append(cmds, Or(
			If(fmt.Sprintf("-f %s", destination), fmt.Sprintf("chmod %o %s", permissions, destination)),
			If(fmt.Sprintf("-d %s", destination), fmt.Sprintf("chmod %o %s/%s", permissions, destination, filenameFromUrl(url))),
			And("echo \"Couldn't determine target\"", "exit 1")))
	}
	return And(cmds...)
}

// Download the URL and extract in the given directory.
func DownloadAndExtract(url, targetDir string) *ShellCommand {
	downloadCmd := download(url)

	if targetDir == "" {
		panic("empty target directory given")
	}

	return And(
		downloadCmd,
		If(fmt.Sprintf("! -d %s", targetDir), Mkdir(targetDir, "", 0)),
		fmt.Sprintf("cd %s", targetDir),
		fmt.Sprintf("tar xvfz /tmp/downloads/%s", filenameFromUrl(url)))
}

func filenameFromUrl(url string) string {
	return path.Base(url)
}

// Wait for the given path to appear, with the given timeout.
func WaitForFile(path string, timeoutInSeconds int) *ShellCommand {
	t := 10 * timeoutInSeconds
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && [ ! -e %s ]; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"file %s did not appear\" 1>&2 && exit 1; }; }",
		t, path, t, path)
	return &ShellCommand{
		Command: cmd,
	}
}

// Wait for the given unix file socket to appear, with the given timeout.
func WaitForUnixSocket(path string, timeoutInSeconds int) *ShellCommand {
	t := 10 * timeoutInSeconds
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && ! { netstat -lx | grep \"%s$\"; }; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"socket %s did not appear\" 1>&2 && exit 1; }; }",
		t, path, t, path)
	return &ShellCommand{
		Command: cmd,
	}
}

func (sc *ShellCommand) Docker(host *host.Host) string {
	return fmt.Sprintf("RUN %s", sc.Command)
}

func (sc *ShellCommand) Shell(host *host.Host) string {
	cmdBuilder := 0

	if sc.isExecutedAsUser() {
		cmdBuilder = 1
	}

	if host.IsSudoRequired() {
		cmdBuilder += 2
	}

	switch cmdBuilder {
	case 0:
		return sc.Command
	case 1:
		return fmt.Sprintf("su -l %s <<EOF\n%s\nEOF\n", sc.user, sc.Command)
	case 2:
		return fmt.Sprintf("sudo bash <<EOF\n%s\nEOF\n", sc.Command)
	case 3:
		return fmt.Sprintf("sudo -- su -l %s <<EOF\n%s\nEOF\n", sc.user, sc.Command)
	}
	panic("should never be reached")
}

func (sc *ShellCommand) Logging(host *host.Host) string {
	s := []string{"[COMMAND]"}

	if host.IsSudoRequired() {
		s = append(s, "[SUDO]")
	}

	if sc.isExecutedAsUser() {
		s = append(s, fmt.Sprintf("[SU:%s]", sc.user))
	}

	s = append(s, fmt.Sprintf(" # %s", sc.Command))

	return strings.Join(s, "")
}

func (sc *ShellCommand) isExecutedAsUser() bool {
	return sc.user != "" && sc.user != "root"
}
