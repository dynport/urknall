// Go subpackage with zwo's command infrastructure. There are some predefined commands (types that implement the
// "Commander" interface), but you can write custom commands, of course. Most commands come with helper functions to
// allow for easy construction when filling runlists.
//
// Its important to understand that most commands are just plain shell commands, that are executed on the host to be
// provisioned. The exception of the rule are commands only required for docker (like the DockerInitCommand for
// example).
package cmd
