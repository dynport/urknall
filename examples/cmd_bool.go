package main

import (
	"fmt"
	"strings"
)

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
