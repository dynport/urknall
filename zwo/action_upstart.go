package zwo

import (
	"fmt"
	"github.com/dynport/dgtk/goup"
	"github.com/dynport/zwo/host"
)

type upstartAction struct {
	upstart *goup.Upstart
	docker  string
	host    *host.Host
}

func (uA *upstartAction) Docker() string {
	if uA.docker != "" {
		return fmt.Sprintf("CMD %s", uA.docker)
	}
	return ""
}

func (uA *upstartAction) Shell() string {
	if uA.upstart == nil {
		return ""
	}
	fA := &fileAction{
		path:    fmt.Sprintf("/etc/init/%s.conf", uA.upstart.Name),
		content: uA.upstart.CreateScript(),
		owner:   "root",
		mode:    0644,
		host:    uA.host,
	}
	return fA.Shell()
}

func (uA *upstartAction) Logging() string {
	return fmt.Sprintf("[UPSTART] Adding upstart script for '%s'.", uA.upstart.Name)
}
