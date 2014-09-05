---
title: Binary
---

# Urknall Binary
{:.no_toc}

Urknall comes in two flavors: binary and library. This is, as there are two
separate problems that need to be solved. On the one hand there must be the
underlying mechanisms to handle the targets, like managing the connection and
executing commands and recording and transfering the output (this is what the
[binary](../binary/) does). On the other hand there must be the tasks that
should be executed. While this is mostly the users' domain making basic stuff
available helps a lot with bootstrapping projects. This is the binary's purpose
and the next subsection is going to explain its evolution. Afterwards the
different use cases are discussed.

* TOC
{:toc}


## Urknall's Evolution

The concepts used matured over the course of about a year when we tested
different approaches and changed direction a few times, until we settled with
something pragmatic and usable. The basic concepts of how the provisioning is
executed have been pretty stable, while being renamed a lot. What we discussed
a lot was the mechanism to handle commands and templates. The problem is that
changing those affects everyone using them, i.e. changing a basic command like
the `FileCommand` (used to write files to the target), that would affect all
usages, as changes break the caching.

We don't want to provide _the_ solution to provisioning certain aspects, but
give you an example. Take it, try it and most important: modify it!


### Fully Integrated Library

First everything was bundled with the library, i.e. the implementations of some
basic commands and templates were part of the library itself. This meant users
would be using them like the following:

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

There are two problem with the approach. If we changed a template or even worse
a command everyone updating urknall would have his caches broken as the
underlying commands changed. This is not acceptable as the changes might break
stuff pretty bad (like urknall trying to reinstall your production database).
And users can't change the implementations easily, i.e. they have to move the
code manually to their own project.

This put a lot of burden on the implementations delivered with urknall, as
implementation couldn't be changed easily. Therefore these artefacts had to be
delivered another way.


### Integrated With A Binary

In a second iteration we added an urknall binary that had the artefacts as
static assets. Users would add these assets to their project as needed. This
had the additional benefit that modification was easy, as the code was already
there. Even updating with later versions was easily possible, i.e. after an
update of the urknall binary new versions could easily be deployed and verified
against the project's version deployed previously.

The downside with this approach is that the binary gets bigger as the number of
assets increases and users must update the urknall binary (and library) to
update commands and templates to the latest version.


### A Binary Accessing Templates On Github

In the last iteration the templates are fetched from github. While this
requires a network connection, it has the benefit, that changes to the
templates require no update of the urknall binary itself.


## Project Scaffolding

The urknall binary can be used to create a simple basic urknall provisioning
tool. This is to help with the basic steps of creating a new project. Besides a
basic file with a `main` function, that initializes and uses the urknall
system, some command definitions (with a fokus on ubuntu based systems) are
added.

	#!shell
	$ urknall init example
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_add_user.go?ref=master"
	saving file "cmd_add_user.go" to "/private/tmp/example/cmd_add_user.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_as_user.go?ref=master"
	saving file "cmd_as_user.go" to "/private/tmp/example/cmd_as_user.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_bool.go?ref=master"
	saving file "cmd_bool.go" to "/private/tmp/example/cmd_bool.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_download.go?ref=master"
	saving file "cmd_download.go" to "/private/tmp/example/cmd_download.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_extract_file.go?ref=master"
	saving file "cmd_extract_file.go" to "/private/tmp/example/cmd_extract_file.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_file.go?ref=master"
	saving file "cmd_file.go" to "/private/tmp/example/cmd_file.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_fileutils.go?ref=master"
	saving file "cmd_fileutils.go" to "/private/tmp/example/cmd_fileutils.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_shell.go?ref=master"
	saving file "cmd_shell.go" to "/private/tmp/example/cmd_shell.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_ubuntu.go?ref=master"
	saving file "cmd_ubuntu.go" to "/private/tmp/example/cmd_ubuntu.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/cmd_wait.go?ref=master"
	saving file "cmd_wait.go" to "/private/tmp/example/cmd_wait.go"
	loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/main.go?ref=master"
	saving file "main.go" to "/private/tmp/example/main.go"

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

There are many more commands. See `godoc` for more information.


## Template Management

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
	available packages:
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


