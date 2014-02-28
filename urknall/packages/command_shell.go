package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// A shell command is just that: something that is executed in a shell on the host to be provisioned. There is quite a
// lot of infrastructure to build such commands. To make construction of complicated commands easier those helpers use
// the most generic type "interface{}". Thereby it is possible to use these functions with "strings" or other
// "ShellCommands" (returned by other helpers for example).
//
// There are some commands that relate to the system's package management. Those are currently based on apt, i.e. only
// debian based systems can be used (our current system of choice is ubuntu server in version 12.04LTS as of this
// writing).
type ShellCommand struct {
	Command string // Command to be executed in the shell.
	user    string // User to run the command as.
}

func (cmd *ShellCommand) Render(i interface{}) {
	cmd.Command = MustRenderTemplate(cmd.Command, i)
}

// Convenience function to run a command as a certain user. Setting an empty user will do nothing, as the command is
// then executed as "root". Note that nested calls will not work. The function will panic if it detects such a scenario.
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

// Combine the given commands with "and", i.e. all commands must succeed. Execution is stopped immediately if one of the
// commands fails, the subsequent ones are not executed! If only one command is given nothing happens.
func And(cmd interface{}, cmds ...interface{}) *ShellCommand {
	cs := mergeSubCommands(cmd, cmds...)

	finalCommand := fmt.Sprintf("{ %s; }", strings.Join(cs, " && "))
	if len(cs) == 1 {
		finalCommand = cs[0]
	}
	return &ShellCommand{Command: finalCommand}
}

// Combine the given commands with "or", i.e. try one after one, untill the first returns success. If only a single
// command is given, nothing happens.
func Or(cmd interface{}, cmds ...interface{}) *ShellCommand {
	cs := mergeSubCommands(cmd, cmds...)

	finalCommand := fmt.Sprintf("{ %s; }", strings.Join(cs, " || "))
	if len(cs) == 1 {
		finalCommand = cs[0]
	}
	return &ShellCommand{Command: finalCommand}
}

func mergeSubCommands(cmd interface{}, cmds ...interface{}) (cs []string) {
	cmdList := make([]interface{}, 0, len(cmds)+1)
	cmdList = append(cmdList, cmd)
	cmdList = append(cmdList, cmds...)

	for i := range cmdList {
		switch cmd := cmdList[i].(type) {
		case *ShellCommand:
			if cmd.user != "" && cmd.user != "root" {
				panic("AsUser not supported in nested commands")
			}
			cs = append(cs, cmd.Command)
		case string:
			if cmd == "" { // ignore empty commands
				panic("empty command found")
			}
			cs = append(cs, cmd)
		default:
			panic(fmt.Sprintf(`type "%T" not supported`, cmd))
		}
	}
	return cs
}

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

// If the tests succeeds run the given command. The test must be based on bash's test syntax (see "man test"). Just
// state what should be given, like for example "-f /tmp/foo", to state that the file (-f) "/tmp/foo" must exist.
//
// Note that this is a double-edged sword, perfectly fit to hurt yourself. Take the following example:
//	[[ -f /tmp/foo ]] && echo "file exists" && exit 1
// The intention is to fail if a certain file exists. The problem is that this doesn't work out. The command must return
// a positive return value if the file does not exit, but it won't. Use the "IfNot" method like in this statement:
//	[[ ! -f /tmp/foo ]] || { echo "file exists" && exit 1; }
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

// If the tests does not succeed run the given command. The tests must be based on bash's test syntax (see "man test").
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

// Extract the file at the given directory. The following file extensions are currently supported (".tar", ".tgz",
// ".tar.gz", ".tbz", ".tar.bz2" for tar archives, and ".zip" for zipfiles).
func ExtractFile(file, targetDir string) *ShellCommand {
	if targetDir == "" {
		panic("empty target directory given")
	}

	var extractCmd *ShellCommand
	switch {
	case strings.HasSuffix(file, ".tar"):
		extractCmd = extractTarArchive(file, targetDir, "")
	case strings.HasSuffix(file, ".tgz"):
		fallthrough
	case strings.HasSuffix(file, ".tar.gz"):
		extractCmd = extractTarArchive(file, targetDir, "gz")
	case strings.HasSuffix(file, ".tbz"):
		fallthrough
	case strings.HasSuffix(file, ".tar.bz2"):
		extractCmd = extractTarArchive(file, targetDir, "bz2")
	case strings.HasSuffix(file, ".zip"):
		extractCmd = &ShellCommand{Command: fmt.Sprintf("unzip -d %s %s", targetDir, file)}
	default:
		panic(fmt.Sprintf("type of file %q not a supported archive", path.Base(file)))
	}

	return And(
		Mkdir(targetDir, "", 0),
		extractCmd)
}

func extractTarArchive(path, targetDir, compression string) *ShellCommand {
	additionalCommand := ""
	switch compression {
	case "gz":
		additionalCommand = "z"
	case "bz2":
		additionalCommand = "j"
	}
	return And(
		fmt.Sprintf("cd %s", targetDir),
		fmt.Sprintf("tar xf%s %s", additionalCommand, path))
}

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

// Wait for the given path to appear. Break and fail if it doesn't appear after the given number of seconds.
func WaitForFile(path string, timeoutInSeconds int) *ShellCommand {
	t := 10 * timeoutInSeconds
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && [ ! -e %s ]; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"file %s did not appear\" 1>&2 && exit 1; }; }",
		t, path, t, path)
	return &ShellCommand{
		Command: cmd,
	}
}

// Wait for the given unix file socket to appear. Break and fail if it doesn't appear after the given number of seconds.
func WaitForUnixSocket(path string, timeoutInSeconds int) *ShellCommand {
	t := 10 * timeoutInSeconds
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && ! { netstat -lx | grep \"%s$\"; }; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"socket %s did not appear\" 1>&2 && exit 1; }; }",
		t, path, t, path)
	return &ShellCommand{
		Command: cmd,
	}
}

func (sc *ShellCommand) Shell() string {
	if sc.isExecutedAsUser() {
		return fmt.Sprintf("su -l %s <<EOF_ZWO_ASUSER\n%s\nEOF_ZWO_ASUSER\n", sc.user, sc.Command)
	}
	return sc.Command
}

func (sc *ShellCommand) Logging() string {
	s := []string{"[COMMAND]"}

	if sc.isExecutedAsUser() {
		s = append(s, fmt.Sprintf("[SU:%s]", sc.user))
	}

	s = append(s, fmt.Sprintf(" # %s", sc.Command))

	return strings.Join(s, "")
}

func (sc *ShellCommand) isExecutedAsUser() bool {
	return sc.user != "" && sc.user != "root"
}
