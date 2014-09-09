---
title: Binary
---

# Urknall Binary
{:.no_toc}

While the urknall [library](../library/) provides the handling of targets,
tasks and caching, the urknall binary helps managing projects that use the
library. This pattern is the result of an evolution that happened over the
course of almost a year. The steps of this evolution and the reasoning behind
them are described in the next subsections. Afterwards different use cases are
discussed.

* TOC
{:toc}


## Urknall's Evolution

The concepts used matured over the course of about a year when different
approaches were tested and direction changed a few times, until development
settled with something pragmatic and usable. The basic concepts of how the
provisioning is executed have been pretty stable, while being renamed a lot.
Discussion mainly focussed on how to handle commands and templates.


### Fully Integrated Library

First the naive approach was used, with bundling everything with the library,
i.e. the implementations of some basic commands and templates were part of the
library itself. This meant users would use them like described in the
following:

	#!golang
	import (
	  "github.com/dynport/urknall/commands"
	  "github.com/dynport/urknall/packages"
	)

	type template {
	  [...]
	}

	func (t *template) Render(pkg urknall.Package) {
	  pkg.AddTemplate("ruby", &packages.Ruby{Version: "2.1.2"})
	  pkg.AddCommands("hello", &commands.Shell("echo hello world"))
	}

With this approach one major problems arises: Changes to templates or (even
worse) commands inside the library would break caches of users after updates.
As commands must not be idempotent this could have fatal consequences and is
not acceptable. As it is not easily possible to fix go library versions when
using them, this could result in different team members using different
versions of a template or command. This is like asking for disaster.

Another issue is that the provided templates are outside the scope of users,
i.e. they can change the templates directly, but have to move the according
code to their project manually. As urknall should not deliver the solutions to
all problems, but at most some initial help, another approach was required,
that allowed users to directly integrate those artefacts into their code base.


### Integrated With A Binary

In a second iteration an urknall binary was added that had the artefacts as
static assets shipped with it. Users would add these assets to their project as
needed. Even updating with later versions was easily possible, under the
assumption the project uses version control, which could be taken for granted
for anything serious. After an update of the urknall binary new versions of
templates could easily be added and changes could be verified against the
project's version using the version control system.

The downside with this approach is that the binary gets bigger as the number of
assets increases and users must update the urknall binary (and library) to
update commands and templates to the latest version.


### A Binary Accessing Templates On Github

In the last iteration the templates are fetched from github. While this
requires a network connection, it has the benefit, that changes to the
templates require no update of the urknall binary itself.


## Usage

Please note that the urknall binary queries the [Github
API](http://github.com){:target='blank'}, that has a rate limit. If you
encouter this, try [creating an API token](https://github.com/blog/1509-personal-api-tokens){:target='blank'}.
The token should reside in the environment as `GITHUB_TOKEN`.

TODO: modify the `init` and `templates add` commands to add files into a
subpackage of the user's project so that godoc can work properly


### Project Scaffolding

The urknall binary can be used to create a simple basic urknall provisioning
tool. This is to help with the basic steps of creating a new project. Besides a
basic file with a `main` function, that initializes and uses the urknall
system, some command definitions (with a fokus on ubuntu based systems) are
added.

	#!shell
	$ urknall init example
	created "cmd_add_user.go"
	created "cmd_as_user.go"
	created "cmd_bool.go"
	created "cmd_download.go"
	created "cmd_extract_file.go"
	created "cmd_file.go"
	created "cmd_fileutils.go"
	created "cmd_shell.go"
	created "cmd_ubuntu.go"
	created "cmd_wait.go"
	created "main.go"

The files installed contain the following commands:

* `InstallPackages` and `UpdatePackages` from `cmd_ubuntu.go`: These commands
  are useful to manage packages on an ubuntu system. Replace these commands
  with something that supports your platform if it is not ubuntu (or debian
  based).
* `Mkdir` from `cmd_fileutils.go`: Create directories with a owner and
  permissions set accordingly.
* `WriteFile` from `cmd_file.go`: The command will write given content to a
  specific files with owner and permissions set as specified. The content will
  be contained in the generated shell command  as gzipped and base64 encoded
  string.
* `Download` and `DownloadAndExtract` from `cmd_download.go`: Uses `curl` to
  download content from a given URI and save or extract respectively it to a
  given location.


### Template Management

We have some basic templates in stock we use a lot. Those can be added to a
project with the urknall binary.

* `urknall templates list` lists all available templates.
* `urknall templates add <template names>` adds the given templates to the
  project.

The `urknall templates list` command lists the available templates. These are
retrieved from urknall's
[github repository](https://github.com/dynport/urknall/tree/master/examples){:target='blank'},
so a network connection is required! At the time when this guide was written
the following templates were available:

	#!shell
	$ urknall templates list
	available templates:
	* docker
	* elasticsearch
	* firewall
	* golang
	* haproxy
	* jenkins
	* nginx
	* openvpn
	* postgis
	* postgres
	* rabbitmq
	* redis
	* ruby
	* syslogng
	* system


