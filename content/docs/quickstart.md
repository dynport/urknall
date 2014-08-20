---
title: QuickStart
layout: default
---

# Getting Started
{:.no_toc}

This guide will help you to create a basic provisioning tool to deploy this
documentation into a fresh host, thereby showing you the nuts and bolts of
urknall.

* TOC
{:toc}


## Requirements

To reproduce the steps of this quickstart guide the following requirements must
be met. First and foremost [go](http://www.golang.org) must be installed and
configured properly. See the documentation provided for details on how to do
that. If you don't have any experience with the go language itselff, then the
[tour of go](http://tour.golang.org/#1) is quite helpful.

When go is up and running you need to install urknall itself. Installation is
as easy as running the following command.

	go get github.com/dynport/urknall/urknall

This will download the source to `$GOPATH/src/github.com/dynport/urknall` and
install the [urknall binary](/binary) to `$GOPATH/bin` (which should be in your
`$PATH` environment for best experience) and [urknall library](/library) to the
subtree in `$GOPATH/pkg`.

This quickstart guide will work with an example that needs to be provisioned. A
fresh Ubuntu Trusty 14.04 based machine should be used to repeat the steps.
Creating a local virtual machine (using something like
[VirtualBox](https://www.virtualbox.org) or [VMWare](http://www.vmware.com)) is
the best option for testing. But using a cloud instance (like from Amazon's
AWS, JiffyBox, etc.) or even bare metal is possible, too. The only requirements
to the target machine are the following:

* The machine must be accessible via SSH, i.e. the SSH server must be running
  and you must know the IP of the machine.
* You must know the username (and if required the password) of a user on the
  machine. If this user is not `root` he must be allowed to run commands using
  `sudo` without being asked for a password, as described in [here](/docs/library/#sudo_without_password).


## Creating The Basic Project

Urknall comes in two parts: a library and a binary. The _library_ provides the
actual mechanisms for provisioning, like creating a connection to the remote
machine, handling the cache that decides which commands must be executed and
executing them.

The _binary_ can be used to create a basic structure of an urknall provisioning
tool in a given directory using the `init` command:

	urknall init example

While not strictly necessary for this simple example you should consider to
stick with go's convention to create source in a directory like
`$GOPATH/src/github.com/<username>/<project>` where `<username>` is your
username on github and `<project>` the name of the project you created. This
will simplify sharing and installing the sources using the go tooling (like `go
get` or `goimports`).

The `urknall init` command creates a lot of files that should be explained
next. The `main.go` will be explained in greater detail below. The `cmd_*.go`
files contain [command](/docs/glossary/#command) definitions. These are
abstractions that allow to specify commands that should be executed on the
remote machine.

~~~ golang
Shell("echo -n hello && echo world")
WriteFile("/tmp/foo", "some content", "root", 0644)
~~~

The first example show a simple shell command, that will be executed directly
on the machine. The second example is much more elaborate, as it will be
expanded into a series of commands that will create a file `/tmp/foo`
containing the line `some content` with owner set to `root` and permissions set
to `0644`. This mechanism has the benefit that code is better readable, the
compiler can support with types and internally logging can be modified to be
more specific on what a series of commands actually does.

The `main.go` file contains the `main` function executed initially. The
relevant part of code for this quickstart guide is in the `run` function, that
initializes urknall, configures the target to be provisioned and finally does
the actual build.

~~~ golang
func run() error {
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
}
~~~

First urknall's logger is configured to use standard output for logging. The
creation and immediate closing of the logger inside the `defer` statement might
seem awkward first, but allows for a pretty concise formulation of the problem.

Next the target is configured using the `urknall.NewSshTarget` or
`urknall.NewSshTargetWithPassword` respectively if public-key authentication is
not usable. Make sure you add the proper values for `uri` consisting of the
username and IP address of the machine you use for this quickstart guide. Also
set the `password` if required.

The last line does two things. A [template](/docs/glossary/#template) is
instantiated and given to urknall's `Run` function, which will render it to the
target built in the previous step. The template is the specification of the
actions to perform on the target. The example template will just `echo` the
famous `hello world` string.

~~~ golang
type Template struct {
}

func (tpl *Template) Render(p urknall.Package) {
  p.AddCommands("hello", Shell("echo hello world"))
}
~~~

Every template must implement the `Renderer` interface. The `Render` method
implemented is given a package that commands are added to. This is where the
commands from all the `cmd_*.go` files come into play. For a detailed
introduction of the commands see the [binary's documentation](/docs/binary).


## Running The Basic Example

After changing the `uri` and `password` variables' value the example can be
compiled and run.

~~~ bash
$ go get . && example
[ubuntu@192.168.56.10:22][hello       ][  0.600][EXEC    ][COMMAND] # echo hello world
[ubuntu@192.168.56.10:22][hello       ][  0.610] + echo hello world
[ubuntu@192.168.56.10:22][hello       ][  0.610] hello world
[ubuntu@192.168.56.10:22][hello       ][  0.621][FINISHED][COMMAND] # echo hello world
~~~

The output shows the command run, the single steps (like in the using the `-x`
flag on `bash`) and a line when the command was finished. This way the exact
runtime can be seen and all intermediate steps a more complicated command may
take. If the example is run a second time the output changes:

~~~ bash
$ go get . && example
[ubuntu@192.168.56.10:22][hello       ][  0.257][CACHED  ][COMMAND] # echo hello world
~~~

This shows the caching mechanism in effect (notice the "CACHED" mark in the
fourth field). As the command was already executed and neither itself nor one
of its (non existing as it is the only command) predecessors changed nothing
had to be done. The next section will show the possibilities of extending the
basic template.


## Extending The Basic Project

The basic template just renders a single `echo` command to the target. To do
something more meaningful the basic example should be extended to provision a
host that serves the [nanoc documentation](http://nanoc.ws/docs/). This
requires the installation of nginx and ruby. Finally the documentation's
repository must be cloned, the static pages be built and nginx be configured to
serve them.

_TODO_: Actually the example should deploy the urknall documentation to the
host, but this requires the repository to be public first.


### The Templating System

For the intended setup _ruby_ and _nginx_ must be installed. While it would be
possible to extend the basic template, this would result in one large template,
that would be hard to read and wouldn't be reusable. Therefore templates can be
added to templates to form hierarchies. Additionally urknall has a mechanism to
retrieve basic templates. These templates might not be exactly what you
require, but could be a good point to start from, i.e. help you to take the
first steps to solve the problem. For a detailed discussion see the
[binary's documentation](/docs/binary#template_management).

The `urknall templates list` command lists the available templates. These are
retrieved from urknall's
[github repository](https://github.com/dynport/urknall/tree/master/examples),
so a network connection is required! At the time when this guide was written
the following templates were available:

~~~ bash
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
~~~

For the software required there are templates available, so it is sufficient to
add those.  The `templates add` command of the `urknall` binary can be used to
download and add a list of templates:

~~~ bash
$ urknall templates add nginx ruby
loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/tpl_nginx.go?ref=master"
loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/tpl_ruby.go?ref=master"
~~~

Now there are two files `tpl_nginx.go` and `tpl_ruby.go` that can be used as
rough sketch for our requirements. You should use a version control system like
[git](http://git-scm.org) and add and commit these templates. This way changes
to the upstream templates can be easily verified by loading the template again
and verify the delta to the local version. As these templates reside locally in
your repository you have maximum flexibility, as you can easily modify them to
suite your needs.

For the quickstart example no modifications are required, so lets inspect the
downloaded files and get a grip on the template mechanism.


### Inspecting The Installed Templates

Every template is a go struct type that implements the `Template` interface,
i.e. it must have a `Render` method that retrieves a `Package` as argument. The
`Package` interface hass three methods to add another template (for the
template hierarchies already mentioned), a task or a list of commands.

Templates have been mentioned already. A _task_ is a list of _commands_ and the
entity that caching applies to (the mechanism that decides whether a command
needs to be executed or not). Manually creating tasks is not required often,
but can be useful if certain commands should only be executed under certain
circumstances. The `AddTask` method's description has an example.

The `AddCommands` method is most important, as it allows to add the actual
commands to be executed in a template. A task will be generated in the
background.

When starting to use a template it's important to verify the template's
configuration possibilities. Those can be found in it's `struct` definition.
Urknall supports a set of annotations that specify whether a value must be
given (is required) or if a default value is given. For the `ruby` package it
look like this:

~~~ golang
type Ruby struct {
	Version string `urknall:"required=true"`
	Local   bool
}
~~~

There is a field `Version` that is required, i.e. must be specified when
the template is instanciated. The other field is optional, as it has no
annotation. It will have the go specific default value, which is `false` for
boolean values.

The nginx template works quite similar. The next subsection will show how to
use these templates.


### Using The Installed Templates

The templates installed in the previous subsections must be integrated into the
template hierarchy to be used. The root of this hierarchy is the template
rendered to the target in `Run` call in the `run` function described in the
beginning of this guide. This value's type will now be modified, to add the
_ruby_ and _nginx_ templates required. For the sake of demonstration the
annotation mechanism to specify a default version for nginx and require an
explicit version for ruby are used (usually both would be set to be required).

~~~ golang
type Template struct {
	RubyVersion  string `urknall:"required=true"`
	NginxVersion string `urknall:"default=1.4.1"`
}
~~~

Next the template's `Render` method is modified to add the templates.
Additionally a command is added to make sure that the system's package
cache is updated and the installed packages are upgraded. This prevents errors
when installing packages fails as the package cache is outdated.

~~~ golang
	p.AddCommands("pkg-update", UpdatePackages())

	p.AddTemplate("ruby", &Ruby{Version: tpl.RubyVersion})
	p.AddTemplate("nginx", &Nginx{Version: tpl.NginxVersion})
~~~

The `AddCommands` and `AddTemplate` are given a string first, that is used to
generate identifiers for the different tasks and used internally to do
bookkeeping on which commands have already been executed. For a deep hierarchy
the segments given with each call to `AddTemplate` are concatenated using a
".".

The last thing to do is specifying the ruby version, that must be given when
instantiating the root template, as it is set as required in the annotations.

~~~ golang
func run() error {
	[...]
	return urknall.Run(target, &Template{RubyVersion: "2.1.2"})
}
~~~

Now the provisioning will update the system's package cache, install upgrades,
ruby and nginx. Still missing are the deployment of the documentation and the
configuration of nginx.


### Further Extending The Templates

Still missing is the actual deployment of the documentation and configuration
of nginx. This requires access to aspects of the already installed templates,
like the path where the stuff was installed.  These are accessible from the
templates itself so its sufficient to keep the values available:

~~~ golang
ruby := &Ruby{Version: tpl.RubyVersion}
nginx := &Nginx{Version: tpl.NginxVersion}

p.AddTemplate("ruby", ruby)
p.AddTemplate("nginx", nginx)
~~~

Now requests to these variables can be issued when doing the actual deployment.

~~~ golang
p.AddCommands("github.docs",
	InstallPackages("git"),
	Shell(ruby.InstallDir()+"/bin/gem install bundle"),
	AsUser("ubuntu", "git clone https://github.com/nanoc/nanoc.ws.git docs"),
	AsUser("ubuntu", "export PATH=${PATH}:"+ruby.InstallDir()+"/bin &&cd docs && bundle install && nanoc compile"),
)
~~~

This will install [git](http://git-scm.org), install the bundle gem, checkout
the documentation repository, install all the gems and finally compile the
pages. Please note that some commands are executed as user `ubuntu` (just for
the sake of showing the feature). The rendered pages will be available in
`/home/ubuntu/docs/output` so the only task remaining is configuring this
directory as root of the nginx server.

~~~ golang
p.AddCommands("nginx.conf",
	Shell(`sed -e "s.root \+html;.root /home/ubuntu/docs/output;." -i `+nginx.ConfDir()+`/nginx.conf`),
	Shell("service nginx start"),
)
~~~

Now everything is setup and configured. The provisioning (again started using
`go get . && example`) will take quite a while was ruby and nginx need to be
compiled. Afterwards everything is set up and served on the virtual machine's
public address (like `http://192.168.56.10`).


## Conclusion

This guide showed how to create a basic provisioning tool for a simple task.

