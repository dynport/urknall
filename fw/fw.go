// Firewall configuration for hosts.
//
// Hosts are configured to have a firewall running, i.e. to only let network traffic pass the interfaces that is allowed
// to. A complete description of firewalling is totally out of scope, but a few things should be mentioned.
//
// The general approach urknall takes, is to disallow everything, but the stuff explicitly allowed. IPTables works by
// letting run packets through tables of rules (different tables for different stages of the network stack; see
// http://en.wikipedia.org/wiki/Iptables for a better description). If a rule hits the end of a table the default policy
// takes effect. This is set to "DROP" in urknall.
//
// IPTables is able to handle the state of a packet, where state is either NEW, ESTABLISHED, or RELATED. For performance
// reasons all packets for the latter two states are accepted (i.e. may pass the firewall) in one of the first rules.
// The remaining state NEW must be granted explicitly for services, hosts, protocols etc.
//
// urknall has a host mode regarding firewall named "paranoid". This is a setting where outgoing traffic must be explicitly
// allowed, too. Like with all paranoia this creates a lot of burdens when doing stuff, but on the other hand it helps
// in some situations. If someone hacks a non "root" account (root could change the firewall rules), he must take
// additional steps to pass the firewall. Maybe that prevents some more damage.
//
// TODO(gfrey): Write about how rules are configured in urknall. Have some nice examples.
package fw
