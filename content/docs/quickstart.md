---
title: Urknall Gettings Started
layout: default
---


# Getting Started

This guide will help you to create a basic provisioning tool (for a simple
example application) and show you the nuts and bolts of urknall.


## Requirements

For testing the tools a target to be built needs to be available. A virtual
machine running Ubuntu 14.04 (that's what this guide is built against) is
sufficient.

Such a VM could be built using VirtualBox or VMWare. Take into account the
following points to prevent problems:

* Make sure you know the username and password of a user on the box. If this
  user is not `root` he must be allowed to run commands using `sudo` without
  being asked for a password (see [this section](#sudo)).
* There must be an SSH server running and accessible from your host. Remember
  the IP address assigned.

It is required that you have the [go](http://www.golang.org) environment
installed and configured on your machine. An introduction to _go_ is out of
scope for this guide. See the linked page's Tour to _go_.

Urknall must be installed first. This can be done using

	go get github.com/dynport/urknall/urknall

Just make sure that the urknall binary `urknall` is located in your `PATH`
environment variable. It can be found in `$GOPATH/bin/`.


## <a name="sudo"></a> Sudo Without Password

Urknall must be able to execute a lot of commands like installing packages or
creating users, which require `root` permissions. If you're not provisioning
using the `root` user the `sudo` mechanism is required. As manual entry of
passwords is tedious it is required that the user is allowed sudo without
password. This can be achieved by adding the following setting (make sure you
change the username from 'ubuntu' to whatever suits you):

	echo "ubuntu ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/90-nopassword

Now verify that there is no password required on running commands with `sudo`.


## Creating The Basic Project

The urknall binary helps with creating the basic structure of a provisioning
project. This can be done using the `init` command of `urknall`

	urknall init $GOPATH/src/github.com/dynport/example

Make sure to replace `dynport` with your github's username. While this is not
that essential for this guide it's good style regarding _go_ applications.

This will create a set of initial files. That should be best added to a git
repository.

	git init . && git add . && git commit -m "initial commit"

In the next section we will have a look what files were generated and how to
use them.


## The Urknall Basic Project And How Run It

Now lets inspect the files that were added by the urknall binary.

* `main.go`: This is the main binary that initializes, configures and runs
             urknall.
* `cmd_*.go`: These are the command definitions. Urknall uses the `command`
              abstraction to run model different commands types.

Let's have closer look at the `main.go` file first and modify the relevant bits
to make the example work.

<pre><code class="language-golang">func run() error {
  defer urknall.OpenLogger(os.Stdout).Close()
  var target urknall.Target
  var e error
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
  return urknall.Run(target, &Template{})
}</code></pre>

The following steps are performed:

* First urknall's logger is configured to use standard output for logging. The
  creation and immediate closing of the logger inside the `defer` statement
  might seem awkward first, but allows for a pretty concise formulation of the
  problem.
* Next the target is configured. Make sure to enter the proper value for `uri`
  and `password`. If no password is specified urknall will try to use
  public-key based login, which is the recommended way of handling login, as no
  password must be stored and communicated.
* The last step is calling urknall's `Run` function, that will render the given
  template `Template` (described next) to the target we built in the previous
  step.

The specification of the actions to perform on the target are described in a
template. The following template will just `echo` the `hello world` string.

<pre><code class="language-golang">type Template struct {
}

func (tpl *Template) Render(p urknall.Package) {
  p.AddCommands("hello", Shell("echo hello world"))
}</code></pre>

Every template must implement the `Renderer` interface. The `Render` method
implemented is given a package that commands are added to. This is where the
commands from all the `cmd_*.go` file come into play. For a detailed
introduction of the commands see the [binary's documentation](/docs/binary).

After changing the `uri` and `password` variables' value you can compile and
run the example:

	go get . && example

The output should look something like this:

	[ubuntu@192.168.56.10:22][hello       ][  0.600][EXEC    ][COMMAND] # echo hello world
	[ubuntu@192.168.56.10:22][hello       ][  0.610] + echo hello world
	[ubuntu@192.168.56.10:22][hello       ][  0.610] hello world
	[ubuntu@192.168.56.10:22][hello       ][  0.621][FINISHED][COMMAND] # echo hello world

Now try to run the binary a second time. Notice the difference in the output:

	[ubuntu@192.168.56.10:22][hello       ][  0.257][CACHED  ][COMMAND] # echo hello world

This shows the caching mechanism in effect. As the command was already executed
and neither itself or noen of its predecessors (there actually are none as it
is the only command there) changed nothing had to be done. Next have look into
the possibilities of extending the basic template.


## Extending The Basic Project

The basic template just renders a single `echo` command to the target. Let's go
and build something more meaningful.
