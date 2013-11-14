// Packages have parameters and compile them into a runlist.
//
// The package's configuration will be used during the compilation of the runlist to faciliate reuse of packages.
package zwo

// A compiler must be able to add commands to a runlist, taking its configuration into account.
type Compiler interface {
	Compile(rl *Runlist) (e error) // Add the package specific commands to the runlist.
}
