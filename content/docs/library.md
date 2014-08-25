---
title: Library
---

# Urknall Library
{:.no_toc}

The library part of urknall is where most of the magic happens. For a detailed
information on the API of urknall have look into the [API
documentation](http://godoc.org/github/dynport/urknall). This guide shows the
basic information required for using urknall.

* TOC
{:toc}


## Commands

What urknall actually does is executing commands on a target. Commands in the
sense of shell commands. Internally these are modelled using the `Commands`
interface. A basic set of implementations is provided using the [urknall
binary](/docs/binary/#urknall_init). There is a most basic `ShellCommand` for
example, that is given a string, that is execute as is. A more advanced example
would be the `FileCommand` that writes given content to file with given owner
and permissions.

The following subsubsection will show different interfaces that must (or can)
be implemented by commands.


### The `Commands` Interface

Every command must implement the `Command` interface:

~~~ golang
type Command interface {
  Shell() string
}
~~~

The `Shell` function must return the command that should be executed on the
remote host, i.e. the plain shell command.


### The `Logger` Interface

Some commands can get pretty complex and obfuscate the real intent by this
complexity. The `FileCommand` already mentioned is an example. To simplify the
logging output, there is the `Logger` interface.

~~~ golang
type Logger interface {
  Logging() string
}
~~~

If a command implements the interface this function is called to generate the
string to be logged. Otherwise the output of the `Shell` function will be used.


### The `Renderer` Interface

When writing templates it's convenient to use it's properties in the command
strings using go's templating (templating in the sense of having special marks
in a string that are replaced with content) mechanism. The following example
show the benefit.

~~~ golang
type ExampleTemplate struct {
  Name string
}

func (et *ExampleTemplate) Render(pkg urknall.Package) {
  pkg.AddCommands("hello", Shell("Hello {{ .Name }}"))
}
~~~

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

~~~ golang
type Renderer interface {
  Render(i interface{})
}
~~~

There is a helper function in the `github.com/dynport/urknall/utils` packages
named `MustRenderTemplate` that can be used to do the actual rendering.


### The `Validator` Interface

The `Validator` interface can be used to do more complex validations, like
making sure all required values are set properly.

~~~ golang
type Validator interface {
  Validate() error
}
~~~

TODO: well that could be described better I guess


## Tasks

Tasks are ordered collections of commands. Usually there is no need to handle
them manually, except for situation's where conditionals are required inside a
cached entity. The following subsection will describe this scenario. In the
following caching will be described as tasks are the layer where it is applied.


### Manual Task Generation

Urknall provides the `NewTask` function that will generate a blank task that
commands can be added to manually.

TODO: missing example that properly explains the need for this. cache breaking for bundle install?


### Caching

One of the core features of urknall is the caching layer that will decide
whether or not a command must be executed. This is essential if provisioning is
run more than once, which is useful in many situations:

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


## Packages

Packages are an strictly internal data-structure. It is a container for all the
tasks that must be executed on the target. The interface is just exposed to the
user when rendering templates. There are three possibilities for adding tasks:

* Using the `AddTemplate` method the given template will be rendered into the
  current template, i.e. all tasks generated inside the "child" template are
  added to the "parent". This is required to build template hierarchies.
* The `AddTask` method will add the given manually created task.
* With `AddCommands` a task is generated internally using the list of commands
  given.

Each of these commands is given a string that is used as identifier for the
underlying task. In case of template hierarchies the different layers' names
are concatenated using dots.


## Templates

Templates are used to define the list of tasks that should be performed during
provisioning. Conceptually they are structs that implement the `Template`
interface, i.e. have a `Render` method that will extend a given `Package`.
These steps are influenced by the configuration of the template.

When building a template hierarchy, from the root template given to the `Build`
function towards some more generic templates it might be necessary to have a
lot of configuration options on the root, that are handed through to the leafs.
This way there is a single interface for setting and changing configuration
which helps with handling more complex scenarios.

TODO: need motivation why we need the special `Template` abstraction layer
instaed of directly using the `Packages`.


### Anonymous Render Function

Sometimes there are templates that don't have any configuration. There is the
`TemplateFunc` mechanism shown in the following example.

~~~ golang
func templateFunc(pkg urknall.Package) {
  pkg.AddCommands("hello", "echo hello world"))
}

type template struct {
  [..]
}

func (t *template) Render(pkg urknall.Package) {
  pkg.Add("anon", urknall.TemplateFunc(templateFunc))
}
~~~

## Targets

The target is the "host" where the commands are executed on. Currently there is
support for remote execution using SSH and running commands locally.


### Remote Target

The remote target mechanism uses SSH to connect to the remote machine and sends
everything back and forth through this channel. The connection opened initially
is kept for the complete session.

There are two basic mechanisms for authentication using SSH, a password or a
public key can be used. They are instantiated using the `NewSshTarget` or
`NewSsshTargetWithPassword` respectively.

Please note that the public key mechanism won't read your `~/.ssh` directory
and you need to add your key to an ssh-agent.


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



