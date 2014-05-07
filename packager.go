package urknall

// A "Package" is an entity that packs commands into a runlist, taking into account their own configuration.
type TaskPackager interface {
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
func NewTask(cmds ...interface{}) *Task {
	return &Task{task: &anonymousTask{cmds: cmds}}
}
