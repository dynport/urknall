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
