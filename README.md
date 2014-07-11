# urknall

urknall is a library written in go for the automated provisioning of servers.

The primary use case for urknall is automated execution of commands on linux (ubuntu) hosts, either locally or via ssh.

urknall does not require any extra software to be installed on the target host (except bash) and has no dependencies. It's designed to be used and shipped as a standalone binary (executable).

urknall __is not a framework__ but a library. There is no (and probably never will be) `urknall provision <my host>` binary. We are also keen on providing building blocks that simplify the life of developers and system operators rather than giving out blueprint-recipes for complex setups that will hardly ever fit their needs completely.

To use urknall you need start with creating a __custom go application__ (either binary or script) and adding the commands which are executed on targets step by step.

## Philosophy

The philosophy behind urknall reflects our design choices while building urknall and improving it over time. We do not deem them to be exclusive or even objective, but it's the principles and ideas our team shares about provisioning and software design.

### Single binary (executable)

All necessary resources (except e.g. credentials) should be compiled into a single, statically linked binary.

Examples:

* configuration templates
* recipes/cookbooks/playbooks/etc.
* template -> target mappings (see Building Blocks)

To allow your teammates to provision a new server you only need to share one single binary. There should be no need to install any software (ruby, python, etc.), libraries (gems, pip packages, etc) or have a checkout of a repository for config templates.

You can (and should) add checks which make sure that you always use the most recent version of the binary (to avoid regressions) by e.g. validating the build revision of the binary with a remote service. Having a single file to maintain reduces the complexity of this chure.


### Type safety
Infrastructure should be describe in a type safe programming language.

* compiler support: catch bugs/typos/etc early (even before doing any provisioning)
* easier refactorings, cleaning up, drying up of your configuration
* reduce duplication and clutter

__YAML is neither type safe, nor a programming language__

Choosing a purely declarative approach or even domain specific languages to infrastructure definition leaves room for interpretation and therefor misunderstanding.

### Pragmatic

urknall Templates (see Buildling Blocks) should not be too generic. They should not include all the nobs to configure your database server. They should also not support multiple linux distributions as you will rarely need to switch your OS platform. If you feel that need, urknall is not the right library for you.
    

## Building Blocks

### Build

Execution of a `template` on a `target`

### Target

Host to be provisioned (remote (via ssh) or local)

### Template

* holds variables
* is rendered to `packages`

### Package

* list of `tasks`

### Task

* list of `commands`
* has unique name (used for Caching)
* unit of caching (see Caching)

### Command

* atomic-ish* linux command

Examples:

* install system packages
* write a file to disk
* download the source of a package

\* e.g. the WriteFile command executes these steps: write file + change owner + change mode

## Caching

urknall uses statement-based caching (similar to building docker images) on a `task` level. 

The checksums of all commands that are executed for a task are written into a tree structure __on the target__. The next time a task is supposed to be executed on a target the current checksums of the task are compared with the checksums of already executed commands on the target. If the checksum of a command changes all commands including that command will be re-executed and the tree is updated. Unchanged commands before the first changed command will not be executed again.

## Hello world

This is the "hello world" of urknall. This program can be either executed with `go run <file.go>` or compiled to a binary `go build -o ./urknall-test <file.go>`.


	package main

	import (
		"log"
		"os"

		"github.com/dynport/urknall"
		"github.com/dynport/urknall/packages"
	)

	func main() {
		if e := provision(); e != nil {
			log.Fatal(e)
		}
	}

	func provision() error {
		// setup logging to stdout
		defer urknall.OpenLogger(os.Stdout).Close()

		// create a basic urknall.Template
		// executes "echo hello world" as user ubuntu on the provided host
		tpl := urknall.TemplateFunc(func(p urknall.Package) {
			p.AddCommands("run", packages.Shell("echo hello world"))
		})

		// create provisioning target for provisioning via ssh with
		// user=ubuntu
		// host=172.16.223.142
		// password=ubuntu
		target, e := urknall.NewSshTargetWithPassword("ubuntu@172.16.223.142", "ubuntu")
		if e != nil {
			return e
		}
		return urknall.Run(target, tpl)
	}

When compiled to a binary this program __does not need any runtime dependencies__ as go statically links all dependencies.

## More Examples

See [examples](examples) for a list of more advanced examples.

