---
title: Urknall Gettings Started
layout: default
---

# Getting Started
{:.no_toc}

This guide will help you to create a basic provisioning tool (for a simple
example application) and show you the nuts and bolts of urknall.

* TOC
{:toc}

## Requirements


For testing the tools a target to be built needs to be available. A virtual
machine running Ubuntu 14.04 (that's what this guide is built against) is
sufficient.

Such a VM could be built using VirtualBox or VMWare. Take into account the
following points to prevent problems:

* Make sure you know the username and password of a user on the box. If this
  user is not `root` he must be allowed to run commands using `sudo` without
  being asked for a password (see [this section](#sudo-without-password)).
* There must be an SSH server running and accessible from your host. Remember
  the IP address assigned.

It is required that you have the [go](http://www.golang.org) environment
installed and configured on your machine. An introduction to _go_ is out of
scope for this guide. See the linked page's Tour to _go_.

Urknall must be installed first. This can be done using

	go get github.com/dynport/urknall/urknall

Just make sure that the urknall binary `urknall` is located in your `PATH`
environment variable. It can be found in `$GOPATH/bin/`.


## Sudo Without Password

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


## The Urknall Basic Project And How To Run It

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
and build something more meaningful. As an example we will deploy nginx hosting
the [nanoc documentation](http://nanoc.ws/docs/). This will require the
installation of nginx and ruby.


### Installing Templates

First we need some templates for ruby and nginx. urknall provides a basic set
of templates. Using the `urknall templates list` command the available
templates are list. Please note this list is assembled using the content of the
`examples` folder of the [github repository](https://github.com/dynport/urknall/tree/master/examples).

This is the currently available list of templates:

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

Luckily there are templates for the required packages, so we just need to add
those to get at least something. The `urknall template add` can be used to
download and add templates:

~~~ bash
$ urknall templ add nginx ruby
loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/tpl_nginx.go?ref=master"
loading content from "https://api.github.com/repos/dynport/urknall/contents/examples/tpl_ruby.go?ref=master"
~~~

Now there are two files `tpl_nginx.go` and `tpl_ruby.go` that can be used as
blueprint for our requirements. You should use a version control system like
[git](http://git-scm.org) and add and commit these templates. This way changes
to the upstream templates can be easily verified by again loading the template
and reviewing the changes to the local version. Now lets inspect the loaded
files and get a grip on the template mechanism.


### Inspecting Installed Templates

Every template is a go struct type that implements the `Template` interface,
i.e. it must have a `Render` method that retrieves a `Package` as argument. The
`Package` interface requires three methods to add another template, a task or a
list of commands.

A *template* is a set of *tasks* and a *task* is a list of *commands*. The
caching is applied per task, i.e. if one of the commands in a task changes,
this and all following commands will be executed in the next run.

As templates can be added while rendering a template hierarchies can be built
easily. This is useful for adding one template multiple times (for example to
configure a dynamic list of users).

The `AddCommands` method is a convenience function as it allows to add a cached
list of commands on the fly. The `AddTask` method can be used if a task must be
created manually. This can be required if some commands must only be executed
based on some condition (there is an example in the methods description).

Let's get started provisioning the ruby template. The struct's field can have
an annotation to mark fields that must be specified or set default values.

This is what the ruby template looks like:

~~~ golang
type Ruby struct {
	Version string `urknall:"required=true"`
	Local   bool
}
~~~

There is a field `Version` that is required, i.e. must be specified when
created. The other field optional, as it has no annotation. It will have the go
specific default value, i.e. `false`.

The nginx template works quite similar. The next subsection will show how to
use these templates.


### Using The Templates

First we'll modify the template we're adding to the build so that the ruby and
nginx template are used. For the sake of demonstration let's use the annotation
mechanism to specify a default version for nginx and require an explicit
version for ruby.

~~~ golang
type Template struct {
	RubyVersion  string `urknall:"required=true"`
	NginxVersion string `urknall:"default=1.4.1"`
}
~~~

Next we'll need to adopt the our template's `Render` method to add the
templates. Additionally we'll add some additional commands to make sure that
the packages installed are updated at least once a day. This demonstrates how
the `Packages.AddCommands` method can be used.

~~~ golang
	timeString := time.Now().UTC().Format("2006-01-02")
	p.AddCommands("base", Shell("# "+timeString), UpdatePackages())
	p.AddTemplate("ruby", &Ruby{Version: tpl.RubyVersion})
	p.AddTemplate("nginx", &Nginx{Version: tpl.NginxVersion})
~~~

Additionally we need to specify the ruby version when instanciating our
template, as it is set as required:

~~~ golang
func run() error {
	[...]
	return urknall.Run(target, &Template{RubyVersion: "2.1.2"})
}
~~~

Now the provisioning will update the apt's package cache, install updates, ruby
and nginx. Still missing are configuration of nginx and the deployment of the
documentation.


### Further Extending The Templates

We're still missing the actual deployment of the documentation. For the
deployment we will need to access aspects of the ruby and nginx
installation, like the insdallation and configuration paths. These are
accessible from the templates itself so lets first keep the values of the types
available:

~~~ golang
ruby := &Ruby{Version: tpl.RubyVersion}
nginx := &Nginx{Version: tpl.NginxVersion}

p.AddTemplate("ruby", ruby)
p.AddTemplate("nginx", nginx)
~~~

Now we can issue requests to these variables when doing the actual deployment.
Next we'll have to get the code, all the required tools and build it finally.

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
pages. Please note that some commands are executed as user `ubuntu`. The
rendered pages will be available in `/home/ubuntu/docs/output` so lets finally
define this directory as root of nginx.

~~~ golang
p.AddCommands("nginx.conf",
	Shell(`sed -e "s.root \+html;.root /home/ubuntu/docs/output;." -i `+nginx.ConfDir()+`/nginx.conf`),
	Shell("service nginx start"),
)
~~~

Now everything is setup and configured. Run the provisioning and afterwards
browse to the server's public address like http://192.168.56.10 and you should
see the documentation.


## Conclusion

Now you've seen how to create a basic provisioning tool. This is just the most
basic example because it lacks support for multiple host provisioning,
deployment of specified versions and many more. But actually that is way beyond
the scope of urknall itself.

