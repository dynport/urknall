# urknall

urknall is a go automation library.

The primary use case for urknall is automated execution of commands on linux (ubuntu) hosts, either locally or with ssh.

urknall does not require any extra software to be installed on the target host (except bash).

urknall __is not a framework__ but a libary. There is no (and probably never will be) a `urknall provision <my host>` binary.

To use urknall you need to create a __custom go application__ (either binary or script) which configures the commands which are executed on targets.

## Philosophie

### Single binary

All necessary resources (except e.g. credentials) should be compiled into a single, statically linked binary.

Examples:

* configuration templates
* recipes/cookbooks/playbooks/etc.
* template -> target mappings (see Building Blocks)

To allow your team to e.g. provision a new server you only need to share a single binary. There should be no need to install any software (ruby, python, etc.), libraries (gems, pip packages, etc) or have a checkout of a repository for e.g. config templates.

You can (and should) add checks which make sure that you always use the most recent version of the binary (to avoid regressions) by e.g. validating the build revision of the binary with a remote service.


### Type safety
Infrastructure should be describe in a type safe porgramming language.

* compiler support: catch bugs/typos/etc early (not when doing the provisioning)
* allows refactorings, cleaning up, drying up of your configuration
* allows reducing of duplications

__YAML is neither type safe, nor a programming language__

### Pragmatic

urknall Templates (see Buildling Blocks) should not be too generic. They should not include all the nobs to e.g. configure your database server. They should also not support multiple linux distributions etc. If you feel that need, urknall is not the right library for you.
    

## Building Blocks

### Build

Execution of a `template` on a `target`

### Target

Remote (via ssh) or local host

### Template

* hold variables
* are rendered to `packages`

### Package

* list of `tasks`

### Task

* list of `commands
* have unique names (used for Caching)
* unit of caching (see Caching)

### Command

* atomic-ish* execution of linux commands

Examples:

* install system packages
* write a file to disk
* download the source of a package

\* e.g. the WriteFile command executes these steps: write file + change owner + change mode

## Caching

urknall uses statement based caching (similar to building docker images) on a `task` level. 

The checksums of all commands which are executed for a task are written into a tree structure __on the target__. The next time a task is executed on a target the current checksums of the task are compared with the checksums on the target. If the checksum of a command changes all commands including that command will be re-executed and the tree is updated. Unchanged commands before the first changed commands will not be executed.

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

