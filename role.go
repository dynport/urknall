package urknall

import "fmt"

// A role is a function that adds packages to a host.
type Role func(host *PackageList)

// A role registry holds information on the available roles and applies all those roles registered for a given host
// (hosts have a list of tags that should have the form `role:<name>` for the roles to be used for it).
type RoleRegistry map[string]Role

// Create a new role registry.
func NewRoleRegistry() *RoleRegistry {
	return &RoleRegistry{}
}

// Add a role with the given name.
func (rr RoleRegistry) Add(name string, role Role) (e error) {
	if _, found := rr[name]; found {
		return fmt.Errorf("Role %q already registered", name)
	}
	rr[name] = role
	return nil
}

// Add a single package as a role (creating the boilerplate for you).
func (rr RoleRegistry) AddPackage(name string, pkg Packager) (e error) {
	return rr.Add(name, func(host *PackageList) { host.Add(name, pkg) })
}

// Apply all matching roles to the given host. The matching criteria is the existence of a host tag of the form
// "role:<rolename>".
//func (rr RoleRegistry) ApplyRoles(host *Host) (e error) {
//	for _, tag := range host.Tags {
//		if strings.HasPrefix(tag, "role:") {
//			name := strings.TrimPrefix(tag, "role:")
//			role, found := rr[name]
//			if !found {
//				return fmt.Errorf("role %q unknown", name)
//			}
//			role(host)
//		}
//	}
//	return nil
//}
