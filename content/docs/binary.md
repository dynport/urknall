---
title: Binary
---

# Urknall Binary

The urknall binary is used to manage urknall based provisioning tools. It is
the last step of a longer "evolution" described in the next subsection


## Urknall's Evolution

The concepts used have matured over the course of about a year when we tested
different concepts and changed directions a few times, until we settled with
something pragmatic and usable. The basic concepts have been pretty stable,
while being renamed a lot. What we discussed a lot was the mechanism to handle
commands and templates. The problem is that changing those affects everyone
using them.

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
  pkg.AddTemplate("ruby", &package.Ruby{Version: "2.1.2"})
  pkg.AddCommands("hello", &commands.Shell("echo hello world"))
}
~~~

There are two problem with the approach. First if we changed a template or even
worse a command everyone updating urknall would have his caches broken as the
underlying commands changed. This is not acceptable as the changes might brake
stuff pretty bad (like urknall trying to reinstall your production database).
Second users can't change the implementations. This puts a lot of burden on the
implementations delivered with urknall. They absolutely have to work for users.
This can't be guaranteed, as there are just to many use cases. Therefore these
artefacts need to be delivered another way.


### Integrated With A Binary

In a second iteration we had the artefacts integrated with the binary as
assets. This was way better than the first approach. Users would import the
templates needed into their project and had all the freedom to change them
according to their needs. Even updating with later versions was easily
possible, i.e. after an update of the urknall binary new versions could easily
be deployed and verified against the version deployed previously.

The downside with this approach is that users must update the urknall binary
(and library) every time they want to update the templates to the latest
version. Which might not actually be wanted.


### A Binary Accessing Templates On Github

In the last iteration the templates are fetched from github. While this
requires a network connection, it has the benefit, that changes to the
templates require no update of the urknall binary itself.


## Project Management

The urknall binary can be used to create a simple basic urknall provisioning
tool. This is to help with the basic steps of creating a new project. Besides a
basic file with a `main` function some command definitions are added.

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

The files added contain some basic commands for usage with ubuntu based
systems (most of them should work with any unix system) and a basic structure
for a provisioning tool that shows the basic urknall initialization and usage.


## Template Management

We have some basic templates in stock we use a lot. Those can be added to a
project with the urknall binary.

* `urknall templates list` will list all available templates.
* `urknall templates add <template names>` will add the given templates to the
  project.


