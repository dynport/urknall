package urknall

import (
	"fmt"
	"strings"
)

type Package struct {
	tasks     []*Task
	taskNames map[string]struct{}
}

func (pkg *Package) Add(name string, sth interface{}) {
	switch v := sth.(type) {
	case *Task:
		v.name = name // safe to set it here
		pkg.addTask(v)
	case *Package:
		pkg.addPackage(name, v)
	case TaskPackager:
		pkg.addTask(&Task{name: name, task: v})
	default:
		panic(fmt.Sprintf("type %T not supported in Package.Add", sth))
	}
}

func (pkg *Package) addPackage(name string, child *Package) {
	pkg.validateTaskName(name)
	for _, task := range child.tasks {
		newTask := &Task{name: name + "." + task.name, task: task.task}
		pkg.addTask(newTask)
	}
}

func (pkg *Package) addTask(task *Task) {
	pkg.validateTaskName(task.name)
	pkg.taskNames[task.name] = struct{}{}
	pkg.tasks = append(pkg.tasks, task)
}

func (h *Package) precompile() (e error) {
	for _, task := range h.tasks {
		if len(task.commands) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", task.name)
		}

		if e = task.compile(); e != nil {
			return e
		}
	}

	return nil
}

func (pkg *Package) validateTaskName(name string) {
	if name == "" {
		panic("package names must not be empty!")
	}

	if strings.Contains(name, " ") {
		panic(fmt.Sprintf(`package names must not contain spaces (%q does)`, name))
	}

	if pkg.taskNames == nil {
		pkg.taskNames = map[string]struct{}{}
	}

	if _, ok := pkg.taskNames[name]; ok {
		panic(fmt.Sprintf("package with name %q exists already", name))
	}
}
