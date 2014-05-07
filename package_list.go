package urknall

import (
	"fmt"
	"strings"
)

type Package struct {
	items []*packageListItem

	packageNames map[string]struct{}
}

type packageListItem struct {
	Key     string
	Package *Task // TODO(gf): rename to something more reasonable!
}

func (pkg *Package) Add(name string, sth interface{}) {
	switch v := sth.(type) {
	case *Task:
		pkg.addTask(name, v)
	case *Package:
		pkg.addPackage(name, v)
	case TaskPackager:
		pkg.addTask(name, &Task{name: name, task: v})
	default:
		panic(fmt.Sprintf("type %T not supported in Package.Add", sth))
	}
}

func (pkg *Package) addPackage(name string, child *Package) {
	pkg.validateTaskName(name)
	for _, item := range child.items {
		itemName := name + "." + item.Key
		item.Package.name = itemName
		pkg.addTask(itemName, item.Package)
	}
}

func (pkg *Package) addTask(name string, task *Task) {
	pkg.validateTaskName(name)
	pkg.packageNames[name] = struct{}{}
	pkg.items = append(pkg.items, &packageListItem{Key: name, Package: task})
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

func (pkg *Package) validateTaskName(name string) {
	if name == "" {
		panic("package names must not be empty!")
	}

	if strings.Contains(name, " ") {
		panic(fmt.Sprintf(`package names must not contain spaces (%q does)`, name))
	}

	if pkg.packageNames == nil {
		pkg.packageNames = map[string]struct{}{}
	}

	if _, ok := pkg.packageNames[name]; ok {
		panic(fmt.Sprintf("package with name %q exists already", name))
	}
}
