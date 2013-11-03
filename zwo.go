// zwo provides everything necessary to provision machines. It provides the mechanisms required to run a set of commands
// on a machine (whatever that is, for example bare metal or a docker container). These commands can be encapsulated in
// a package that provides configuration so that reuse is possible. Every package adds the raw commands to a runlist for
// a dedicated host. This allows for provisioning different packages to one host. A caching mechanism is used, so that
// repeated provisioning of the same package will only run the commands necessary (a changed command and all
// subsequent).
//
//  TODO: Explain commands (just plain stupid shell commands).
//  TODO: Talk about idempotent commands.
//  TODO: How is a package configured and what is done using reflection?
//  TODO: How are runlists added to the mix?
//  TODO: How is provisioning different targets (ssh vs. docker) done?
//  TODO: Talk about configuration file handling.
package zwo
