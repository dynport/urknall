package zwo

import (
	"fmt"
	"github.com/dynport/dgtk/goup"
	"github.com/dynport/zwo/assets"
	"github.com/dynport/zwo/host"
	"github.com/dynport/zwo/utils"
	"os"
)

// A runlist is a container for commands. While those can have arbitrary intent, they should be closely related, for the
// sake of clarity and reusability.
type Runlist struct {
	actions []action
	host    *host.Host
	config  interface{}
	name    string // Name of the compilable.
}

func (rl *Runlist) ExecuteAsUser(user, command string) {
	if user == "" || user == "root" {
		panic(fmt.Errorf("user must be given and not be root (was '%s')", user))
	}
	cmd := rl.createCommandForExecute(command)
	cmd.user = user
	rl.actions = append(rl.actions, cmd)
}

func (rl *Runlist) Execute(command string) {
	cmd := rl.createCommandForExecute(command)
	rl.actions = append(rl.actions, cmd)
}

func (rl *Runlist) WaitForFile(path string, timeoutInSeconds int) {
	t := 10 * timeoutInSeconds
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && [ ! -e %s ]; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"file %s did not appear\" 1>&2 && exit 1; }; }",
		t, path, t, path)
	rl.actions = append(rl.actions, &commandAction{cmd: cmd, host: rl.host})
}

func (rl *Runlist) WaitForSocket(path string, timeoutInSeconds int) {
	t := 10 * timeoutInSeconds
	cmd := fmt.Sprintf(
		"x=0; while ((x<%d)) && ! { netstat -lx | grep \"%s$\"; }; do x=\\$((x+1)); sleep .1; done && { ((x<%d)) || { echo \"socket %s did not appear\" 1>&2 && exit 1; }; }",
		t, path, t, path)
	rl.actions = append(rl.actions, &commandAction{cmd: cmd, host: rl.host})
}

func (rl *Runlist) createCommandForExecute(command string) (c *commandAction) {
	if command == "" {
		panic("empty command given")
	}

	renderedCommand := utils.MustRenderTemplate(command, rl.config)
	return &commandAction{cmd: renderedCommand, host: rl.host}
}

func (rl *Runlist) AddFile(path, assetName, owner string, mode os.FileMode) {
	if path == "" {
		panic("no path given")
	}

	asset, e := assets.Get(assetName)
	if e != nil {
		panic(fmt.Errorf("error retrieving asset: %s", e.Error()))
	}

	content := utils.MustRenderTemplate(string(asset), rl.config)
	rl.actions = append(rl.actions, &fileAction{path: path, content: content, owner: owner, mode: mode, host: rl.host})
}

func (rl *Runlist) Init(us *goup.Upstart, ds string) {
	if us == nil && ds == "" {
		panic("neither upstart nor docker run command given")
	}

	rl.actions = append(rl.actions, &upstartAction{upstart: us, docker: ds, host: rl.host})
}

// The configuration is used to expand the templates used for the commands, i.e. all fields and methods of the given
// entity are available in the template string (using the common "{{ .Something }}" notation, see text/template for more
// information).
func (rl *Runlist) setConfig(cfg interface{}) {
	rl.config = cfg
}

// For the caching mechanism a unique identifier for each runlist is required. This identifier is set internally by the
// provisioner.
func (rl *Runlist) setName(name string) {
	rl.name = name
}

func (rl *Runlist) getName() (name string) {
	return rl.name
}
