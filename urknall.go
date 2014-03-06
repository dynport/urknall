// urknall is a different approach towards the provisioning problem. It is based on the experience that the generic
// solutions available don't actually work. Everything generic must compromise here and there. It also requires taking
// into account a multitude of problems and thereby getting pretty complex soon. And with all these compromises and
// complexity comes the problem of being highly error prone and not being quite right anyway.
//
// urknall is different as it is a set of tools to easily set up and maintain tools for provisioning machines. It helps
// to bootstrap such a tool easily, but gets out the way as much as possible. This is done by having a library that will
// manage everything required to actually execute stuff on a machine, adding a layer of "caching" that will only do what
// needs to be done, and having some decent logging. The other part is a binary that will help you bootstrap a new
// project (adding template code to a new directory) and adding template commands and packages (those will be explained
// soon).
//
// The problem with the generic solution problem is solved using different tools. First everything is written in golang,
// a language with pretty decent tooling like a blazing fast compiler, working "go to definition" tools for many
// editors, and many more. On the other hand the templates provided are just that templates. As they reside in the
// user's project he can adopt them as much as he likes (using "rpm" instead of "deb" for package management for
// example).
//
// The core building block of urknall is the `Host` datastructure that defines the machine to provision. It has
// information how to access the machine, but also the knowledge what to actually do on it. This latter is added using
// commands and packages.
//
// A command is just a simple bash command executed on the remote host. A package is a set of commands combined with
// some configuration which allows for reuse of the same commands, for example to install different versions of the same
// program. Commands are collected in runlists. One runlist per package or command registered at the host. It must be
// given an unique name. Internally urknall will maintain the state of the runlist's commands already run. Only if a
// command hasn't been executed yet, or one of the predecessors changed, it will be run. This is a simple but very
// effective caching mechanism. Additionally urknall has the possibility to display which commands would actually be
// executed, i.e. a dry run.
package urknall

import "sync"

func Provision(host *Host) (e error) {
	prov := newProvisioner(host, nil)
	return prov.provision()
}

func ProvisionMulti(hosts ...*Host) (elist []error) {
	wg := &sync.WaitGroup{}

	eChannel := make(chan error, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(wg *sync.WaitGroup, eChannel chan<- error, host *Host) {
			defer wg.Done()

			prov := newProvisioner(host, nil)

			if e := prov.provision(); e != nil {
				eChannel <- e
			}
		}(wg, eChannel, host)
	}

	wg.Wait()

	for e := range eChannel {
		elist = append(elist, e)
	}

	return elist
}

func ProvisionDryRun(host *Host) (e error) {
	prov := newProvisioner(host, &provisionOptions{DryRun: true})
	return prov.provision()
}
