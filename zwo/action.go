package zwo

import (
	"fmt"
	"github.com/dynport/zwo/host"
	"github.com/dynport/zwo/templates"
	"os"
	"strings"
)

// The action interface must be implemented by the different action types.
type action interface {
	Docker() string  // Used for executing the action in a docker context.
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Plain() string   // Used for internal purposes like combining actions.
	Logging() string // Get string used for logging.
}

// The actionFunc type is used to allow easy but error save command chaning. Reasoning behind this concept is, that the
// added complexity in here (handling functions) allows for more elegant creation of runlists (the containers of
// actions), without to much error handling clutter.
type actionFunc func(h *host.Host, iface interface{}) (c action, e error)

// The CommandActionFunc is the basic type (actually a function value) that is used to create command actions, that can
// be added to a runlist using the Runlist_AddCommands method.
//
// The command strings are rendered using text/template and the package to be compiled in order to make these strings
// easier to read.
type CommandActionFunc func(h *host.Host, iface interface{}) (c *commandAction, e error)

// The FileActionFunc is the basic type (actually a function value) that is used to create actions that will write
// files. These functions create actions that can be added to a runlist using the Runlist_AddFiles method.
type FileActionFunc func(h *host.Host, iface interface{}) (c *fileAction, e error)

// The Execute action is the most basic command action and just runs the provided string.
func Execute(cmd string) CommandActionFunc {
	return func(h *host.Host, iface interface{}) (c *commandAction, e error) {
		if cmd == "" {
			return nil, fmt.Errorf("empty command given")
		}
		rendered, e := templates.RenderTemplateFromString(cmd, iface)
		if e != nil {
			return nil, e
		}
		return &commandAction{cmd: rendered, host: h}, nil
	}
}

// Install the given packages using apt.
//
// TODO: This limits usage of the package to debian based systems, but this should be sufficient in most cases.
func InstallPackages(pkgs ...string) CommandActionFunc {
	return func(h *host.Host, iface interface{}) (c *commandAction, e error) {
		if len(pkgs) == 0 {
			return nil, fmt.Errorf("empty package list given")
		}
		cmd := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", strings.Join(pkgs, " "))
		rendered, e := templates.RenderTemplateFromString(cmd, iface)
		if e != nil {
			return nil, e
		}
		return &commandAction{cmd: rendered, host: h}, nil
	}
}

// Execute the given list of actions. It operates in a short circuit fashion, i.e. if one subcommand fails the
// subsequent ones wont be executed. Grouping commands is currently important as the maximum number of commands is
// limited for usage with docker containers (maximum number of AUFS layers is limited to about 42).
func And(cmds ...CommandActionFunc) CommandActionFunc {
	return func(h *host.Host, iface interface{}) (c *commandAction, e error) {
		if len(cmds) == 0 {
			return nil, fmt.Errorf("empty list of commands given")
		}
		if len(cmds) == 1 {
			return cmds[0](h, iface)
		}
		cmdString, e := mergeCommands(h, iface, cmds...)
		if e != nil {
			return nil, e
		}
		return &commandAction{cmd: cmdString, host: h}, nil
	}
}

// If the given test (see "man test") succeeds, execute the given commands.
func If(test string, cmds ...CommandActionFunc) CommandActionFunc {
	return func(h *host.Host, iface interface{}) (c *commandAction, e error) {
		if test == "" {
			return nil, fmt.Errorf("empty test given")
		}
		if len(cmds) == 0 {
			return nil, fmt.Errorf("empty list of commands given")
		}

		cmdString, e := mergeCommands(h, iface, cmds...)
		if e != nil {
			return nil, e
		}
		renderedTest, e := templates.RenderTemplateFromString(test, iface)
		if e != nil {
			return nil, e
		}
		cmd := fmt.Sprintf("test %s && { %s }", renderedTest, cmdString)
		return &commandAction{cmd: cmd, host: h}, nil
	}
}

// If the given test (see "man test") fails, execute the given commands.
func IfNot(test string, cmds ...CommandActionFunc) CommandActionFunc {
	return func(h *host.Host, iface interface{}) (c *commandAction, e error) {
		if test == "" {
			return nil, fmt.Errorf("empty test given")
		}
		if len(cmds) == 0 {
			return nil, fmt.Errorf("empty list of commands given")
		}
		cmdString, e := mergeCommands(h, iface, cmds...)
		if e != nil {
			return nil, e
		}
		renderedTest, e := templates.RenderTemplateFromString(test, iface)
		if e != nil {
			return nil, e
		}
		cmd := fmt.Sprintf("test %s || { %s }", renderedTest, cmdString)
		return &commandAction{cmd: cmd, host: h}, nil
	}
}

func mergeCommands(h *host.Host, iface interface{}, cmds ...CommandActionFunc) (mergedCommand string, e error) {
	cmdStrings := make([]string, 0, len(cmds))
	for i := range cmds {
		c, e := cmds[i](h, iface)
		if e != nil {
			return "", e
		}
		cmdStrings = append(cmdStrings, c.Plain())
	}
	return strings.Join(cmdStrings, " && "), nil
}

// Write the given content to the file at the given path. Set owner and mode accordingly.
func WriteFile(path, content, owner string, mode os.FileMode) FileActionFunc {
	return func(h *host.Host, iface interface{}) (c *fileAction, e error) {
		if path == "" {
			return nil, fmt.Errorf("empty path given")
		}
		if content == "" {
			return nil, fmt.Errorf("empty content given")
		}
		return &fileAction{path: path, content: content, owner: owner, mode: mode, host: h}, nil
	}
}

// Run the given commands as a given user.
//
// TODO: Make sure we don't end in hell because of wrongly nested quotes (encoding?).
func AsUser(user string, cmds ...CommandActionFunc) CommandActionFunc {
	return func(h *host.Host, iface interface{}) (c *commandAction, e error) {
		if user == "" {
			return nil, fmt.Errorf("empty user given")
		}
		cmdString, e := mergeCommands(h, iface, cmds...)
		if e != nil {
			return nil, e
		}
		return &commandAction{cmd: cmdString, host: h, user: user}, nil
	}
}
