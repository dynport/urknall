package zwo

// A 'Compiler' is an entity (lets call it a package) that adds commands to a given runlist, taking into account their
// own configuration.
type Compiler interface {
	Compile(rl *Runlist) // Add the package specific commands to the runlist.
}
