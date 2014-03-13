package urknall

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dynport/urknall/cmd"
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

	rlStack rlStack
}

// Get the user used to access the host. If none is given the "root" account is used as default.
func (h *Host) user() string {
	if h.User == "" {
		return "root"
	}
	return h.User
}

// Add a new top level runlist ....
func (h *Host) Add(name string, sth interface{}) {
	h.rlStack.Push(name)
	defer h.rlStack.Pop()

	switch val := sth.(type) {
	case string:
		h.add(NewPackage(val))
	case cmd.Command:
		h.add(NewPackage(val))
	case Package:
		h.add(val)
	case Role:
		val.Apply(h)
	default:
		if reflect.ValueOf(val).Kind() != reflect.Ptr {
			panic(fmt.Sprintf("value %T not a pointer (see http://golang.org/doc/faq#different_method_sets)", val))
		}
		panic(fmt.Sprintf("invalid type (doesn't implement Command, Package, or Role interface): %T", val))
	}
}

// Add the given package with the given name to the host.
//
// The name is used as reference during provisioning and allows for provisioning the very same package in different
// configuration (with different version for example). Package names must be unique.
func (h *Host) add(pkg Package) {
	name := h.rlStack.String()

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
