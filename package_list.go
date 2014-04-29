package urknall

import (
	"fmt"
	"strings"
)

type Package struct {
	items []*PackageListItem

	packageNames map[string]struct{}
}

type PackageListItem struct {
	Key     string
	Package *Task
}

// Alias for the AddCommands methods.
func (h *Package) Add(name string, cmd interface{}, cmds ...interface{}) {
	h.AddCommands(name, cmd, cmds...)
}

// Register the list of given commands (either of the cmd.Command type or as string) as a package (without
// configuration) with the given name.
func (h *Package) AddCommands(name string, cmd interface{}, cmds ...interface{}) {
	cmdList := append([]interface{}{cmd}, cmds...)
	h.AddPackage(name, NewPackage(cmdList...))
}

// Add the given package with the given name to the host.
//
// The name is used as reference during provisioning and allows for provisioning the very same package in different
// configuration (with different version for example). Package names must be unique and the "uk." prefix is reserved for
// urknall internal packages.
func (h *Package) AddPackage(name string, pkg Packager) {
	if strings.HasPrefix(name, "uk.") {
		panic(fmt.Sprintf(`package name prefix "uk." reserved (in %q)`, name))
	}

	if strings.Contains(name, " ") {
		panic(fmt.Sprintf(`package names must not contain spaces (%q does)`, name))
	}

	if h.packageNames == nil {
		h.packageNames = map[string]struct{}{}
	}

	if _, ok := h.packageNames[name]; ok {
		panic(fmt.Sprintf("package with name %q exists already", name))
	}

	h.packageNames[name] = struct{}{}
	packager := &Task{name: name, pkg: pkg}
	h.items = append(h.items, &PackageListItem{Key: name, Package: packager})
}

func (h *Package) precompileRunlists() (e error) {
	for _, item := range h.items {
		if len(item.Package.commands) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", item.Key)
		}

		if e = item.Package.compile(); e != nil {
			return e
		}
	}

	return nil
}
