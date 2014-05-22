package main

import "github.com/dynport/urknall"

type Jenkins struct {
	Plugins    []string
	Version    string `urknall:"default=1.551"`
	HeapSize   string `urknall:"default=512m"`
	ListenAddr string `urknall:"default=127.0.0.1"`
	ListenPort string `urknall:"default=8080"`
}

func (pkg *Jenkins) Render(r urknall.Package) {
	urlRoot := "http://mirrors.jenkins-ci.org"
	url := urlRoot + "/war/" + pkg.Version + "/jenkins.war"

	r.AddCommands("base",
		InstallPackages("openjdk-6-jdk", "bzr", "git"),
		DownloadToFile(url, "/opt/jenkins.war", "root", 0644),
		Mkdir("/data/jenkins", "root", 0755),
		Mkdir("/data/jenkins/plugins", "root", 0755),
		WriteFile(
			"/etc/init/jenkins.conf",
			jenkinsUpstart,
			"root", 0644))

	for _, plugin := range pkg.Plugins {
		pSrc := urlRoot + "/plugins/" + plugin + "/latest/" + plugin + ".hpi"
		pDest := "/data/jenkins/plugins/" + plugin + ".hpi"
		r.AddCommands("plugin-"+plugin, DownloadToFile(pSrc, pDest, "root", 0644))
	}
}

const jenkinsUpstart = `
start on runlevel [23456]
stop on runlevel [!$RUNLEVEL]

env JENKINS_HOME=/data/jenkins
env LANG=en_US.UTF-8

exec /usr/bin/java -Xmx{{ .HeapSize }} -Xms{{ .HeapSize }} -jar /opt/jenkins.war --httpPort={{ .ListenPort }} --httpListenAddress={{ .ListenAddr }} 1>&1 | logger -i -t jenkins
`
