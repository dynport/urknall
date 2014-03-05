package urknall

import (
	"fmt"
	"strings"
)

// The host type, used to describe a host and define everything that should be provisioned on it.
type Host struct {
	IP       string // Host's IP address used to provision the system.
	User     string // User used to log in.
	Password string // SSH password to be used (besides ssh-agent)
	Port     int    // SSH Port to be used

	Tags []string // Tags can be used to trigger certain actions (this is used for the role concept for example).
	Env  []string // Custom environment settings to be used for all sessions.

	packageNames []string
	runlists     []*Runlist
}

// Get the user used to access the host. If none is given the "root" account is used as default.
func (h *Host) user() string {
	if h.User == "" {
		return "root"
	}
	return h.User
}

// Alias for the AddCommands methods.
func (h *Host) Add(name string, cmds ...interface{}) {
	h.AddCommands(name, cmds...)
}

// Register the list of given commands (either of the cmd.Command type or as string) as a package (without
// configuration) with the given name.
func (h *Host) AddCommands(name string, cmds ...interface{}) {
	h.AddPackage(name, NewPackage(cmds...))
}

// Add the given package with the given name to the host.
//
// The name is used as reference during provisioning and allows for provisioning the very same package in different
// configuration (with different version for example). Package names must be unique.
func (h *Host) AddPackage(name string, pkg Package) {
	if strings.Contains(name, " ") {
		panic(fmt.Sprintf(`package names must not contain spaces (%q does)`, name))
	}

	for i := range h.packageNames {
		if h.packageNames[i] == name {
			panic(fmt.Sprintf("package with name %q exists already", name))
		}
	}

	h.packageNames = append(h.packageNames, name)
	h.runlists = append(h.runlists, &Runlist{name: name, pkg: pkg, host: h})
}

// Provision the host, i.e. execute all the commands contained in the packages registered with this host.
func (h *Host) Provision(opts *ProvisionOptions) (e error) {
	sc := newSSHClient(h, opts)
	return sc.provision()
}

// Predicate to test whether sudo is required (user for the host is not "root").
func (h *Host) isSudoRequired() bool {
	if h.User != "" && h.User != "root" {
		return true
	}
	return false
}

func (h *Host) precompileRunlists() (e error) {
	for _, runlist := range h.runlists {
		if len(runlist.commands) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", runlist.name)
		}

		if e = runlist.compile(); e != nil {
			return e
		}
	}

	return nil
}
