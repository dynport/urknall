package main

import "fmt"

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
