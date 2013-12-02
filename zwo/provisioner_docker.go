package zwo

import (
	"fmt"
	"github.com/dynport/dgtk/dockerclient"
	"github.com/dynport/gologger"
	"github.com/dynport/zwo/host"
	"strings"
)

type dockerClient struct {
	baseImage  string
	tag        string
	host       *host.Host
	dockerHost *dockerclient.DockerHost

	dockerfile string
}

func newDockerClient(host *host.Host) (client *dockerClient, e error) {
	dh, e := dockerclient.NewViaTunnel(host.IPAddress(), host.User())
	if e != nil {
		return nil, e
	}
	if host.Docker.WithRegistry {
		dh.Registry = host.IPAddress() + ":5000"
	}
	return &dockerClient{host: host, dockerHost: dh}, nil
}

func (dc *dockerClient) provisionImage(tag string, packages ...Compiler) (imageId string, e error) {
	logger.PushPrefix(dc.host.IPAddress())
	defer logger.PopPrefix()

	if packages == nil || len(packages) == 0 {
		e := fmt.Errorf("compilables must be given")
		logger.Errorf(e.Error())
		return "", e
	}

	if tag != "" {
		if !strings.Contains(tag, "/") && dc.dockerHost.Registry != "" {
			tag = dc.dockerHost.Registry + "/" + tag
		}
		dc.tag = tag
	}

	runLists, e := precompileRunlists(dc.host, packages...)
	if e != nil {
		return "", e
	}

	aLen := countActions(runLists)
	if aLen >= 42 {
		return "", fmt.Errorf("docker only supports runlists with up to 42 commands (found %d)", aLen)
	}

	if dc.baseImage == "" {
		dc.baseImage = "ubuntu"
	}
	dc.dockerfile = fmt.Sprintf("FROM %s\n", dc.baseImage)

	// Provisioning the runlist actually means building a dockerfile.
	if e = provisionRunlists(runLists, dc.buildDockerFile); e != nil {
		return "", e
	}

	// Use the generated dockerfile to build the image.
	imageId, e = dc.dockerHost.BuildImage(dc.dockerfile, dc.tag, logger.Writer(gologger.DEBUG))
	if e != nil {
		return "", e
	}

	if dc.tag != "" {
		e = dc.dockerHost.PushImage(dc.tag, nil)
	}

	return imageId, e
}

func countActions(runLists []*Runlist) (i int) {
	for idx := range runLists {
		i += len(runLists[idx].actions)
	}
	return i
}

func (dc *dockerClient) buildDockerFile(rl *Runlist) (e error) {
	for i := range rl.actions {
		dc.dockerfile += rl.actions[i].Docker() + "\n"
	}
	return nil
}
