package urknall

import (
	"fmt"
	"strings"
)

type PackageList struct {
	Items []*PackageListItem

	packageNames []string
	userRunlists []*Package
}

type PackageListItem struct {
	Key     string
	Package *Package
}

// Add the given package with the given name to the host.
func (list *PackageList) addSystemPackage(name string, pkg Packager) (e error) {
	name = "uk." + name
	for i := range list.packageNames {
		if list.packageNames[i] == name {
			return fmt.Errorf("package with name %q exists already", name)
		}
	}

	list.packageNames = append(list.packageNames, name)
	return nil
}

// Alias for the AddCommands methods.
func (h *PackageList) Add(name string, cmd interface{}, cmds ...interface{}) {
	h.AddCommands(name, cmd, cmds...)
}

// Register the list of given commands (either of the cmd.Command type or as string) as a package (without
// configuration) with the given name.
func (h *PackageList) AddCommands(name string, cmd interface{}, cmds ...interface{}) {
	cmdList := append([]interface{}{cmd}, cmds...)
	h.AddPackage(name, NewPackage(cmdList...))
}

// Add the given package with the given name to the host.
//
// The name is used as reference during provisioning and allows for provisioning the very same package in different
// configuration (with different version for example). Package names must be unique and the "uk." prefix is reserved for
// urknall internal packages.
func (h *PackageList) AddPackage(name string, pkg Packager) {
	if strings.HasPrefix(name, "uk.") {
		panic(fmt.Sprintf(`package name prefix "uk." reserved (in %q)`, name))
	}

	if strings.Contains(name, " ") {
		panic(fmt.Sprintf(`package names must not contain spaces (%q does)`, name))
	}

	for i := range h.packageNames {
		if h.packageNames[i] == name {
			panic(fmt.Sprintf("package with name %q exists already", name))
		}
	}

	h.packageNames = append(h.packageNames, name)
	h.userRunlists = append(h.userRunlists, newRunlist(name, pkg, nil))
}
