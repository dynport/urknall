---
title: Binary
---

# Urknall Binary

Urknall comes in two flavors: binary and library. This is, as there are two
separate problems that need to be solved. On the one hand there must be the
underlying mechanisms to handle the targets, like managing the connection and
executing commands and recording and transfering the output (this is what the
[binary](/docs/binary) does). On the other hand there must be the tasks that
should be executed. While this is mostly the users' domain making basic stuff
available helps a lot with bootstrapping projects. This is the binary's purpose
and the next subsection is going to explain its evolution. Afterwards the
different use cases are discussed.


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

~~~ golang
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
~~~

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


## Project Management

The urknall binary can be used to create a simple basic urknall provisioning
tool. This is to help with the basic steps of creating a new project. Besides a
basic file with a `main` function, that initializes and uses the urknall
system, some command definitions (with a fokus on ubuntu based systems) are
added.

~~~ shell
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
~~~


## Template Management

We have some basic templates in stock we use a lot. Those can be added to a
project with the urknall binary.

* `urknall templates list` lists all available templates.
* `urknall templates add <template names>` adds the given templates to the
  project.


