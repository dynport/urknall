package cmd

import (
	"fmt"
	"github.com/dynport/dgtk/goup"
	"github.com/dynport/zwo/host"
)

// Create an upstart command. That is a script executed on system start. See the github.com/dynport/dgtk/goup package
// for further details.
type UpstartCommand struct {
	Upstart *goup.Upstart // Upstart configuration.
}

func (uA *UpstartCommand) Docker(host *host.Host) string {
	return ""
}

func (uA *UpstartCommand) Shell(host *host.Host) string {
	if uA.Upstart == nil {
		return ""
	}
	fA := &FileCommand{
		Path:        fmt.Sprintf("/etc/init/%s.conf", uA.Upstart.Name),
		Content:     uA.Upstart.CreateScript(),
		Permissions: 0644,
	}
	return fA.Shell(host)
}

func (uA *UpstartCommand) Logging(host *host.Host) string {
	return fmt.Sprintf("[UPSTART] Adding upstart script for '%s'.", uA.Upstart.Name)
}
