# urknall

## Example
    
    host := &urknall.Host{
    	IP:       "127.0.0.1",
    	Hostname: "my-urknall-host",
    	User:     "root", // root is default, if != root all commands are run with sudo
    }
    
    // run commands, "upgrade" is the name which is used to cache execution
    host.AddCommands("upgrade", "apt-get update", "apt-get uprade -y")
    
    // write files
    host.AddCommands("marker", cmd.WriteFile("/tmp/installed.txt", "OK", "root", 0644))
    
    // install packages (implementing urknall.Package)
    host.AddPackage("nginx", nginx.New("1.4.4"))
    
    // provision host with ssh and no extra options
    // make sure your ssh key(s) are loaded (ssh-add -l)
    host.Provision(nil)

For a more complex configuration see e.g. [github.com/dynport/urknall/blob/master/urknall-example/rack_app.go](https://github.com/dynport/urknall/blob/master/urknall-example/rack_app.go)

## Provisioner
The main provisioner is using SSH to connect to the host in question and run the required actions. There is one other
provisioner though, that reuses the packages (collections of commands) and provisions docker container images.


## Hosts
Hosts are the basic structure urknall works with. A host is a target if provisioned itself, but could also be the build
environment for docker container images (and host those in a private registry).


## Packages
Packages are the entities of urknall, that allow to have structure and modularity in the stuff provisioned to hosts. While
it would be totally be possible to just write a list of commands to a host, that wouldn't be easy to manage, maintain
and reuse. Packages should be as small as possible (like providing the commands to set up one programm or service), but
as large as necessary (there is no sense in putting each and every command into a separate package).

Packages have two purposes: one the one hand they are the reusable container for configuration (what version of ruby
should be installed for example), on the other hand they modify the internal datastructure `Runlist`. The latter is done
implementing the `Compiler` interface. The required method `Compile` is given such a Runlist that can be filled with
commands. During provisioning these commands are first compiled (taking the configuration of the package and host into
account) and then executed. Albeit there is a cache, that will prevent commands from being executed, as long as all
predecessor are unchanged (if a command needs to be exeucted as it changed compared to the previous run, all successors
are executed too).


## Requirements
The following subsections should be taken into account when provisioning systems.


### SSH
Some debian systems don't have a root account you can directly log on too, but require to use sudo. This is kind of
tedious in some situations and makes matters difficult for server provisioning. But there is one benefit: while there
are constant attacks against certain well known account names (like root, www, news, ...) normal user accounts are
highly unlikely to be hit by a request. Together with only allowing ssh to certain accounts (like the set of
administrators) and only allowing ssh login via public key authentication the system should be pretty secure even with
ssh being available.

There are some requirements though. All administrators must be in the sudoers group and sudo should be
configured to not
ask for passwords. This can be done editing the sudoers file (use the `sudo visudo` command to get there). The following
line should be modified:
	%sudo ALL=(ALL:ALL) ALL
to be like
	%sudo ALL=(ALL:ALL) NOPASSWD:ALL


### Docker Images
Building docker container images requires a host that is running docker and has the build flag set. This is required to
set some firewall rules to make the internet (where base images will be loaded for example) accessible.

