package zwo

import (
	"fmt"
	"github.com/dynport/dgtk/goup"
	"github.com/dynport/zwo/assets"
	"github.com/dynport/zwo/host"
	"github.com/dynport/zwo/utils"
	"os"
)

// A runlist is a container for commands. Use the following methods to add new commands.
type Runlist struct {
	actions []action
	host    *host.Host
	config  interface{}
	name    string // Name of the compilable.
}

// Execute the given command as the given user (aka su).
func (rl *Runlist) ExecuteAsUser(user, command string) {
	if user == "" || user == "root" {
		panic(fmt.Errorf("user must be given and not be root (was '%s')", user))
	}
	cmd := rl.createCommandForExecute(command)
	cmd.user = user
	rl.actions = append(rl.actions, cmd)
}

// Execute the given shell command.
func (rl *Runlist) Execute(command string) {
	cmd := rl.createCommandForExecute(command)
	rl.actions = append(rl.actions, cmd)
}

func (rl *Runlist) createCommandForExecute(command string) (c *commandAction) {
	if command == "" {
		panic("empty command given")
	}

	renderedCommand := utils.MustRenderTemplate(command, rl.config)
	return &commandAction{cmd: renderedCommand, host: rl.host}
}

// Add the asset with the given name at the path with owner and permission set accordingly.
func (rl *Runlist) AddAsset(path, assetName, owner string, mode os.FileMode) {
	asset, e := assets.Get(assetName)
	if e != nil {
		panic(fmt.Errorf("error retrieving asset '%s': %s", assetName, e.Error()))
	}
	rl.AddFile(path, string(asset), owner, mode)
}

// Add the file wth the given content at the path with owner and permission set accordingly.
func (rl *Runlist) AddFile(path, content, owner string, mode os.FileMode) {
	if path == "" {
		panic("no path given")
	}

	c := utils.MustRenderTemplate(content, rl.config)
	rl.actions = append(rl.actions, &fileAction{path: path, content: c, owner: owner, mode: mode, host: rl.host})
}

// Create upstart script (or docker start command respectively).
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
