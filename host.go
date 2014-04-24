package urknall

import "fmt"

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
	IP       string // Host's IP address used to provision the system.
	User     string // User used to log in.
	Password string // SSH password to be used (besides ssh-agent)
	Port     int    // SSH Port to be used

	Tags []string
	Env  []string // custom env settings to be used for all sessions
}

// Get the user used to access the host. If none is given the "root" account is as default.
func (h *Host) user() string {
	if h.User == "" {
		return "root"
	}
	return h.User
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

		if e = runlist.compile(); e != nil {
			return e
		}
	}

	return nil
}
