---
title: QuickStart
layout: default
---

# Getting Started
{:.no_toc}

This guide will demonstrate how to create a provisioning tool to deploy this
documentation's pages behind a nginx, thereby showing the nuts and bolts of
urknall.

* TOC
{:toc}


## Requirements

Make sure urknall is properly installed as described [here](../../#installation).
Additionally a machine is needed to provision to (see
[VirtualBox](https://www.virtualbox.org){:target='blank'}  on how to create a virtual machine),
that should meet the following requirements:

* The machine should have Ubuntu Trust 14.04 installed.
* The machine must be accessible via SSH.
* If the user on the machine is not `root`, he must be allowed to run commands
  using `sudo`, [without being asked for a password](../library/#sudo_without_password).


## Creating The Basic Project

First the basic project is created in the `example` subdirectory, using the
urknall [binary](../binary/)'s `init` command:

	#!shell
	$ urknall init example

Besides the a basic set of [commands](../binary/#project-scaffolding), the
`main.go` file is creatted, that ahs code that configures and drives the
urknall [library](../library) with a simple [template](../library/#templates).
The generated code can be compiled to a binary, used to do the actual
provisioning.


## Running The Basic Example

After changing the `uri` and `password` variables' values in the `main.go` file
the example code can be compiled and run.

	#!bash
	$ go get . && example
	[ubuntu@192.168.56.10:22][hello       ][  0.600][EXEC    ][COMMAND] # echo hello world
	[ubuntu@192.168.56.10:22][hello       ][  0.610] + echo hello world
	[ubuntu@192.168.56.10:22][hello       ][  0.610] hello world
	[ubuntu@192.168.56.10:22][hello       ][  0.621][FINISHED][COMMAND] # echo hello world

The output shows a log of all commands run and their respective outputs (as
seen when using bash's `-x` flag). If the example is run a second time the
output changes:

	#!bash
	$ go get . && example
	[ubuntu@192.168.56.10:22][hello       ][  0.257][CACHED  ][COMMAND] # echo hello world

This shows the caching mechanism in effect (notice the _CACHED_ mark in the
fourth field). As the command was already executed and nothing changed, nothing
had to be done. The next section will show the possibilities of extending the
basic template.


## Extending The Basic Project

The basic template instanciated when calling urknall is now extended to
demonstrate the execution of proper commands. The task at hand is to serve
urknall's documentation using nginx. Ruby is required to generate the pages
using the [nanoc](http://nanoc.ws){:target='blank'} static page generator.


### The Templating System

First the templates for _ruby_ and _nginx_ provided with urknall are installed.
The urknall binary is able to [list](../binary/#template_management) and
[add](../binary/#template_management) provided templates:

	#!bash
	$ urknall templates list
	available packages:
	[..]
	* nginx
	* ruby
	[..]
	$ urknall templates add nginx ruby
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/tpl_nginx.go?ref=master"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/tpl_ruby.go?ref=master"

Now there are two files `tpl_nginx.go` and `tpl_ruby.go` that can be used as
rough sketch for this project's requirements. Actually no modifications to
these templates are required for this project, so next they can be included
with the provisioning mechanism.


### Using The Installed Templates

The templates installed in the previous subsections must be integrated into the
template hierarchy to be used. The root of this hierarchy is the template
rendered to the target with the call to urknall's `Run` method. This value's
type is now modified, to add the _ruby_ and _nginx_ templates required. For
the sake of demonstration the [annotation mechanism](../library/#annotations)
is used to specify a default version for nginx and require an explicit version
for ruby.

	#!golang
	type Template struct {
	  RubyVersion  string `urknall:"required=true"`
	  NginxVersion string `urknall:"default=1.4.1"`
	}

Next the template's `Render` method is modified. First a command is added to
make sure that the system's package cache is updated and the installed packages
are upgraded. This prevents errors related to an outdated package cache. Next
the two templates that were added in the peviosn chapter are added.

	#!golang
	func (tpl *Template) Render(p urknall.Package) {
	  p.AddCommands("pkg-update", UpdatePackages())

	  ruby := &Ruby{Version: tpl.RubyVersion}
	  p.AddTemplate("ruby", ruby)

	  nginx := &Nginx{Version: tpl.NginxVersion}
	  p.AddTemplate("nginx", nginx)
	}

In a second step the root template is configured to specify the required ruby
version.

	#!golang
	func run() error {
	  [...]
	  return urknall.Run(target, &Template{RubyVersion: "2.1.2"})
	}

The final tasks retrieve the documentation's code, compile it and configure
nginx to serve the output.

	#!golang
	func (tpl *Template) Render(p urknall.Package) {
	  [...]
	  // Clone the documentation and compile it using nanoc.
	  p.AddCommands("github.docs",
		InstallPackages("git"),
		Shell(ruby.InstallDir()+"/bin/gem install bundle"),
		AsUser("ubuntu", "git clone https://github.com/nanoc/nanoc.ws.git docs"),
		AsUser("ubuntu", "export PATH=${PATH}:"+ruby.InstallDir()+"/bin &&cd docs && bundle install && nanoc compile"),
	  )
	
	  // Configure nginx to use nanoc's output as root.
	  p.AddCommands("nginx.conf",
		Shell(`sed -e "s.root \+html;.root /home/ubuntu/docs/output;." -i `+nginx.ConfDir()+`/nginx.conf`),
		Shell("service nginx start"),
	  )
	}

Now everything is setup, configured and started, i.e. the documentation is
reachable using the machine's public IP.


## Conclusion

This guide showed how to create a basic provisioning tool for the simple task
of deploying some static pages generated with a static page generator. The
final code should look like the following, with the commands and templates
being unaltered.

	#!golang
	package main

	import (
	  "log"
	  "os"

	  "github.com/dynport/urknall"
	)

	var logger = log.New(os.Stderr, "", 0)

	func main() {
	  if e := run(); e != nil {
	    logger.Fatal(e)
	  }
	}

	type Template struct {
	  RubyVersion  string `urknall:"required=true"`
	  NginxVersion string `urknall:"default=1.4.1"`
	}

	func (tpl *Template) Render(p urknall.Package) {
	  p.AddCommands("pkg-update", UpdatePackages())

	  ruby := &Ruby{Version: tpl.RubyVersion}
	  p.AddTemplate("ruby", ruby)

	  nginx := &Nginx{Version: tpl.NginxVersion}
	  p.AddTemplate("nginx", nginx)

	  // Clone the documentation and compile it using nanoc.
	  p.AddCommands("github.docs",
		InstallPackages("git"),
		Shell(ruby.InstallDir()+"/bin/gem install bundle"),
		AsUser("ubuntu", "git clone https://github.com/nanoc/nanoc.ws.git docs"),
		AsUser("ubuntu", "export PATH=${PATH}:"+ruby.InstallDir()+"/bin && cd docs && bundle install && nanoc compile"),
	  )

	  // Configure nginx to use nanoc's output as root.
	  p.AddCommands("nginx.conf",
		Shell(`sed -e "s.root \+html;.root /home/ubuntu/docs/output;." -i `+nginx.ConfDir()+`/nginx.conf`),
		Shell("service nginx start"),
	  )
	}

	func run() error {
	  defer urknall.OpenLogger(os.Stdout).Close()
	  var target urknall.Target
	  var e error
	  uri := "ubuntu@192.168.56.10"
	  password := "ubuntu"
	  if password != "" {
		target, e = urknall.NewSshTargetWithPassword(uri, password)
	  } else {
		target, e = urknall.NewSshTarget(uri)
	  }
	  if e != nil {
		return e
	  }
	  return urknall.Run(target, &Template{RubyVersion: "2.1.2"})
	}

