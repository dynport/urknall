package urknall

import (
	"fmt"

	"github.com/dynport/urknall/fw"
)

// The host type. Use the "NewHost" function to create the basic value.
//
// Please note that you need to set the primary interface (the one the host is accessible on) name, if that is not
// "eth0". That should only be necessary on rare circumstances.
//
// A host is added a set of packages, that are provisioned on request.
//
//	TODO(gfrey): Add better support for interfaces and IPs.
//	TODO(gfrey): Add handling and support for IPv6 (currently the firewall will block everything).
type Host struct {
	IP        string // Host's IP address used to provision the system.
	User      string // User used to log in.
	Password  string // SSH password to be used (besides ssh-agent)
	Port      int    // SSH Port to be used
	Hostname  string // Hostname used on the system.
	Interface string // Primary network interface of the host.
	Timezone  string // Local Timezone to be set

	Tags []string
	Env  []string // custom env settings to be used for all sessions

	Paranoid bool // Make the firewall as restrictive as possible.
	WithVPN  bool // Connect host to a VPN. Assumes "tun0" as interface.

	BuildHost               bool   // Whether this host should be used to build urknall binary packages.
	BinaryPackageRepository string // Where should urknall binary packages be searched?

	Firewall fw.Firewall // List of rules used for the firewall.
	IPSets   []*fw.IPSet // List of ipsets for the firewall.

}

// Get the user used to access the host. If none is given the "root" account is as default.
func (h *Host) user() string {
	if h.User == "" {
		return "root"
	}
	return h.User
}

// Get the host's primary interface. If none is given "eth0" is returned as default.
func (h *Host) publicInterface() string {
	if h.Interface == "" {
		return "eth0"
	}
	return h.Interface
}

// Provision the host, i.e. execute all the commands contained in the packages registered with this host.
//func (h *Host) Provision(opts *ProvisionOptions) (e error) {
//	sc := newSSHClient(h, opts)
//	return sc.provision()
//}

// Create a binary package from the given package. This is an optimization for packages that download, compile and
// install sources. As compilation might take its time this step can be done once and the effort be reused. Keep in mind
// that:
//
//	* Packages to be precompiled should only contain the steps necessary to compile and install the sources
//	  (configuration will only be executed during the package creation).
//	* Packages must implement the extend BinaryPackage interface.
//	* There must be a binary package repository reachable and configured for all hosts that should use the binary
//	  packages.
//	* The host to build binary packages on must have the BuildHost flag set. This is to make sure that the side effects
//	  (like installed compilers, extracted sources, etc) are known to the administration and don't happen by accident.
func (h *Host) BuildBinaryPackage(pkg BinaryPackage) (e error) {
	if !h.BuildHost {
		return fmt.Errorf("Host %q is not a build host.", h.Hostname)
	}
	sc := newSSHClient(h, nil)
	return sc.buildBinaryPackage(pkg)
}

// Predicate to test whether sudo is required (user for the host is not "root").
func (h *Host) isSudoRequired() bool {
	if h.User != "" && h.User != "root" {
		return true
	}
	return false
}

func (h *PackageList) runlists() (r []*Package) {
	r = make([]*Package, 0, len(h.userRunlists))
	r = append(r, h.userRunlists...)
	return r
}

func (h *PackageList) precompileRunlists() (e error) {
	for _, runlist := range h.runlists() {
		if len(runlist.commands) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", runlist.name)
		}

		if e = runlist.compileWithBinaryPackages(); e != nil {
			return e
		}
	}

	return nil
}
