package zwo

import (
	"fmt"
	"github.com/dynport/dpgtk/docker_client"
	"github.com/dynport/zwo/host"
	"runtime/debug"
	"strings"
)

type dockerClient struct {
	baseImage string
	tag       string
	host      *host.Host
	client    *docker_client.Client
	image     *docker_client.Image
}

func (dc *dockerClient) Provision(packages ...Compiler) (e error) {
	if packages == nil || len(packages) != 1 {
		return fmt.Errorf("the zwo docker client only supports a single package")
	}

	pkg := packages[0]
	rl := &Runlist{host: dc.host}
	rl.setConfig(pkg)
	rl.setName(getPackageName(pkg))

	pkg.Compile(rl)
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to compile: %v", r)
			}
			logger.Info(e.Error())
			logger.Debug(string(debug.Stack()))
		}
	}()

	if len(rl.actions) >= 42 {
		return fmt.Errorf("docker only supports runlists with up to 42 commands (found %d)", len(rl.actions))
	}

	if e = dc.provision(rl); e != nil {
		return fmt.Errorf("failed to provision: %s", e.Error())
	}

	return nil
}

func (dc *dockerClient) provision(rl *Runlist) (e error) {
	dockerFile := dc.buildDockerFile(rl)

	imageId, e := dc.client.Build(dc.tag, dockerFile, func(s string) { logger.Info(s) })
	if e != nil {
		return e
	}

	dc.image = &docker_client.Image{Client: dc.client, Id: imageId}

	return dc.tagImage()
}

func (dc *dockerClient) buildDockerFile(rl *Runlist) string {
	if dc.baseImage == "" {
		dc.baseImage = "ubuntu"
	}
	lines := []string{"FROM " + dc.baseImage}
	for i := range rl.actions {
		lines = append(lines, rl.actions[i].Docker())
	}
	return strings.Join(lines, "\n")
}

func (dc *dockerClient) tagImage() (e error) {
	if dc.tag != "" {
		dc.image.Repository = dc.tag
		e = dc.image.FetchDetails()
		if e != nil {
			return e
		}
		created, e := dc.image.ImageDetails.CreatedAt()
		if e != nil {
			return e
		}
		e = dc.client.Tag(dc.image.Id, dc.tag, created.UTC().Format("2006-01-02T150405"))
		if e != nil {
			return e
		}
		return dc.client.Tag(dc.image.Id, dc.tag, "latest")
	}
	return nil
}
