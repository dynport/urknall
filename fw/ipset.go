package fw

import (
	"fmt"
	"net"
)

// IPSets are the possibility to change a rule, without actually rewriting the rules. That is they add some sort of
// flexibility with regard to dynamic entities like a load balancer, which must have access to the different machines
// that should take the load.
//
// A set is defined by a name (that is used in a rule, see "Rule.(Source|Destination).IPSet"), that is written to the
// rules. The type defines what parameters must match an entry (see "ipset --help" output and the man page for a list of
// allowed values).
type IPSet struct {
	Name     string   // Name of the ipset, used as reference in the iptables rules.
	Type     string   // Type is something like "hash:ip" or "hash:ip,port".
	Family   string   // Network address family (either "inet" or "inet6"). Defaults to "inet".
	HashSize int      // Size of the hash. You shouldn't need to change this parameter. Defaults to 1024.
	MaxElem  int      // Number elements the set can have at most. Defaults to 65536.
	Members  []net.IP // Initial set of members to add to the set.
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
