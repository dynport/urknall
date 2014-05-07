package urknall

// A "Package" is an entity that packs commands into a runlist, taking into account their own configuration.
type Tasker interface {
	Package(*Task) // Add the package specific commands to the runlist.
}

type anonymousTask struct {
	cmds []interface{}
}

func (anon *anonymousTask) Package(pkg *Task) {
	for i := range anon.cmds {
		pkg.Add(anon.cmds[i])
	}
}

// Create a package from a set of commands.
func NewTask(cmds ...interface{}) Tasker {
	return &anonymousTask{cmds: cmds}
}

// Initialize the given struct reading, interpreting and validating the 'urknall' annotations given with the type.
func InitializePackage(pkg interface{}) error {
	return validatePackage(pkg)
}
