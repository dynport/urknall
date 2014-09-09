---
title: Library
---

# Urknall Library
{:.no_toc}

The library part of urknall provides the core mechanisms to execute commands on
a target. For a detailed information on the API of urknall have look into the
[API documentation](http://godoc.org/github/dynport/urknall){:target='blank'}. This guide
explains the basic ideas behind the concepts.

* TOC
{:toc}


## Commands

What urknall actually does is executing commands on a target. Commands in the
sense of shell commands. Internally these are modelled using the `Commands`
interface. A basic set of implementations is provided using the [urknall
binary](../binary/#urknall_init). There is a most basic `ShellCommand` for
example, that is given a string, that is executed as is. A more advanced
example would be the `FileCommand` that writes given content to a file with
given owner and permissions.

The following subsubsections will show different interfaces that must (or can)
be implemented by commands and their intent.


### The `Commands` Interface

Every command must implement the `Command` interface:

	#!golang
	type Command interface {
	  Shell() string
	}

The `Shell` function must return the command that should be executed on the
remote host, i.e. the plain shell command. All standard `sh` features are
supported, from pipes to subshells and more.

TODO: add information on how to (or better not) use input and output redirects.


### The `Logger` Interface

Some commands can get pretty complex and obfuscate the real intent by this
complexity. The `FileCommand` already mentioned is an example. To simplify the
logging output, there is the `Logger` interface.

	#!golang
	type Logger interface {
	  Logging() string
	}

If a command implements this interface the function is called to generate the
string used for logging. Otherwise the raw output of the `Shell` function will
be used.


### The `Renderer` Interface

When writing templates it's convenient to use it's properties in the command
strings using go's templating (templating in the sense of having special marks
in a string that are replaced with content) mechanism. The following example
show the benefit.

	#!golang
	type ExampleTemplate struct {
	  Name string
	}

	func (et *ExampleTemplate) Render(pkg urknall.Package) {
	  pkg.AddCommands("hello", Shell("Hello {{ .Name }}"))
	}

This way no complex string concatenation is required, but values and functions
can be used directly. Error detection is deferred from compile to run time, but
as command building happens prior to starting the actual execution urknall will
fail early.

There are commands where the rendering must be limited to specific parts and it
is not sufficient to just render the output of the `Shell` function. This is a
problem for example with the the `FileCommand` example where the given content
must be rendered, as it encoded (base64) and zipped when returned.

For this to work commands must be rendered prior to usage. The `Renderer`
interface shows this is supported.

	#!golang
	type Renderer interface {
	  Render(i interface{})
	}

There is a helper function in the `github.com/dynport/urknall/utils` packages
named `MustRenderTemplate` that can be used to do the actual rendering.


### The `Validator` Interface

The `Validator` interface can be used to do more complex validations, like
making sure all required values are set properly.

	#!golang
	type Validator interface {
	  Validate() error
	}

TODO: well that could be described better I guess


## Packages

Packages are an strictly internal data-structure. A package is a container for
tasks that must be executed on the target. The interface is just exposed to the
user when rendering [templates](#templates) and provides three functions:

* Using the `AddTemplate` method the given template will be rendered into the
  current template, i.e. all tasks generated inside the "child" template are
  added to the "parent". This is required to build template hierarchies.
* The `AddTask` method will add the given manually created task.
* With `AddCommands` a task is generated internally using the list of commands
  given.

Each of these commands is given a string that is used as identifier for the
underlying task. In case of template hierarchies the different layers' names
are concatenated using dots.


## Tasks

Tasks are ordered collections of commands. They are the unit caching is applied
to. Caching is one of the core features of urknall, that decides whether a
command must be executed or not, i.e. whether the command and its predecessor
have already been executed. This is essential if provisioning is run more than
once, which is useful in many situations:

* While developing templates the turnaround time is pretty short, as only
  changed or added parts need to be executed.
* When an already provisioned target must be extended only the relevant parts
  are touched.
* Changes to an existing setup are possible without disrupting the overall
  service.

Without this feature repeated provisioning would only be possible if all
commands would be idempotent, i.e. could be run over and over again without
changing results. This is a stark restriction that would require a lot of
thought to get right.

TODO: Is it even possible to build proper idempotent commands without being
      restricted to trivial problems?

Each task is defined by the ordered list of commands that need to be executed.
The commands are identified by the hash of the command actually executed. If it
was executed successfully a file will be written on the target. These files can
be found under `/var/lib/urknall/<task-name>/<hash>.done`. Prior to running a
task all files with this pattern from this directory will be listed.  This list
can be used as foundation for the cache. If a command's hash is contained in
this list the command must not be executed again. If it isn't all remaining
commands must be executed.

TODO: add information on how to best partition the cache.


## Templates

Templates are used to define the list of tasks that should be performed during
provisioning. These tasks are added to a package provided on the `Render`
method call of the `Template`
[interface](http://godoc.org/github.com/dynport/urknall#Template){:target='_blank'},

When building a template hierarchy, from the root template given to the `Build`
function towards some more generic templates it might be necessary to have a
lot of configuration options on the root, that are handed through to the leafs.
This way there is a single interface for setting and changing configuration
which helps with handling more complex scenarios.

When creating the `main.go` file using the [urknall binary](../binary/#init) a
simple template without configuration is generated.

	#!golang
	type Template struct {
	}
	
	func (tpl *Template) Render(p urknall.Package) {
	  p.AddCommands("hello", Shell("echo hello world"))
	}

As the template has no configuration the example could be changed to use the
`TemplateFunc` mechanism as described in the next subsection. The last
subsection describes the annotation mechanism provided by urknall to give
constraints on a template's configuration.


### Anonymous Render Function

Sometimes there are templates that don't have any configuration. There is the
`TemplateFunc` mechanism shown in the following example to avoid unnecessary code.

	#!golang
	func templateFunc(pkg urknall.Package) {
	  pkg.AddCommands("hello", Shell("echo hello world"))
	}

	func run() {
	  return urknall.Run(target, urknall.TemplateFunc(templateFunc))
	}


### Annotations

Urknall provides an annotation based mechanism to give further constraints on a
template's configuration. In the [quickstart guide](../quickstart/) the
`required` and `default` tags were used.

	#!golang
	type Template struct {
	  RubyVersion  string `urknall:"required=true"`
	  NginxVersion string `urknall:"default=1.4.1"`
	}

Prior to rendering templates to a target, urknall will validate it. This
validation takes the annotations into account, i.e. it verifies that:

* `required` fields have not go's zero value set.
* fields with a `default` tag get this value set if none was specified.
* an integer field with `min` or `max` annotations fullfill the respective
  constraints.
* for string fields the value's length is validate if the `size` annotation is
  given.

This helps to prevent missing configuration items prior to executing commands,
that would fail otherwise.


## Logging

Urknall's logging must be configured prior to usage. Internally a
publisher-subscribe mechanism is used, that has more complex features, but the
default configuration should be sufficient in most cases. The `main.go` file
created by the [urknall binary](../binary/#init) does so in the first line of
the run function:

	#!golang
	func run() error {
	  defer urknall.OpenLogger(os.Stdout).Close()
	  // [...]
	}

This configures the logging to write all output to the process's standard
output channel and close the logger on program termination (this is done using
the `defer` statement).

TODO: add more content on how to add a custom logger.


## Targets

The target is the "host" where the commands are executed on. Currently there is
support for remote execution using SSH and running commands locally.


### Remot Target

The remote target mechanism uses SSH to connect to the remote machine and sends
everything back and forth through this secured channel. The connection opened
initially is kept for the complete session.

Authentication on the remote host is either done using a password or
public key for the used user. The password based approach shouldn't be used for
production setups so, but might be the most pragmatic solution for testing
purposes.

The public key authentication mechanism doesn't search the `~/.ssh` directory
for keys, but relies on a configure _ssh-agent_ running.

The `main.go` file created by the [urknall binary](../binary/#init) provides
both mechanisms depending on the availability of a password, using the
`NewSshTarget` and `NewSshTargetWithPassword` respectively. The `uri` and
`password` must be configured depending of the user's use case, of course.

	#!golang
	func run() error {
	  // [...]
	  var target urknall.Target
	  uri := "ubuntu@my.host"
	  password := ""
	  if password != "" {
	    target, e = urknall.NewSshTargetWithPassword(uri, password)
	  } else {
	    target, e = urknall.NewSshTarget(uri)
	  }
	  if e != nil {
	    return e
	  }
	 // [...]
	}


### Local Target

This target can be used to provision the local host.

TODO: motivation?


### Sudo Without Password

Urknall must be able to execute commands like installing packages or creating
users, which require `root` permissions. If you're not provisioning
using the `root` user the `sudo` mechanism is required. As manual entry of
passwords is tedious it is required that the user is allowed sudo without
password. This can be achieved by adding the following setting (make sure you
change the username from 'ubuntu' to whatever suits you):

	echo "ubuntu ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/90-nopassword

Now you should verify that there is no password required on running commands
with `sudo`.


