package fw

import (
	"fmt"
	"net"
)

// IPSets are the possibility to change a rule, without actually rewriting the rules. That is they add some sort of
// flexibility with regard to dynamic entities like a load balancer, which must have access to the different machines
// that should take the load.
//
// A set is defined by a name, that is used in iptables rule (see "Rule.(Source|Destination).IPSet") to reference the
// contained entities. The type defines what parameters must match an entry (see "ipset --help" output and the man page
// for a list of allowed values), for example a set could define hosts and ports.
//
// The family defines the type of IP address to handle, either IPv4 or IPv6. The allowed values are "inet" and "inet6"
// respectively.
//
// There are some ipset internal parameters that shouldn't need to be changed often. Those are "HashSize" that defines
// the size of the underlying hash. This value defaults to 1024. The "MaxElem" number determines how much elements there
// can be in the set at most.
//
// An initial set of members can be defined, if reasonable.
type IPSet struct {
	Name     string   // Name of the ipset.
	Type     string   // Type of the ipset.
	Family   string   // Network address family.
	HashSize int      // Size of the hash.
	MaxElem  int      // Max number of elements of the set.
	Members  []net.IP // Initial set of members.
}

func (ips *IPSet) IPSetRestore() (cmd string) {
	cmd = fmt.Sprintf("create %s %s family %s hashsize %d maxelem %d\n",
		ips.Name, ips.Type, ips.family(), ips.hashsize(), ips.maxelem())
	for idx := range ips.Members {
		cmd += fmt.Sprintf("add %s %s\n", ips.Name, ips.Members[idx].String())
	}
	return cmd + "\n"
}

func (ips *IPSet) family() string {
	if ips.Family == "" {
		return "inet"
	}
	return ips.Family
}

func (i *IPSet) hashsize() int {
	if i.HashSize == 0 {
		return 1024
	}
	return i.HashSize
}

func (i *IPSet) maxelem() int {
	if i.MaxElem == 0 {
		return 65536
	}
	return i.MaxElem
}
