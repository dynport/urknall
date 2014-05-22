package urknall

import (
	"fmt"
	"strings"
	"github.com/dynport/urknall/utils"
)

type PackageBuilder interface {
	BuildPackage(pkg Package)
}

type Package interface {
	Add(string, interface{})
	AddTask(Task)
	Tasks() []Task
}

type packageImpl struct {
	tasks          []Task
	taskNames      map[string]struct{}
	reference      interface{} // used for rendering
	cacheKeyPrefix string
}

func (p *packageImpl) Tasks() []Task {
	return p.tasks
}

func (pkg *packageImpl) Add(name string, sth interface{}) {
	if pkg.cacheKeyPrefix != "" {
		name = pkg.cacheKeyPrefix + "." + name
	}
	name = utils.MustRenderTemplate(name, pkg.reference)
	switch v := sth.(type) {
	case *taskImpl:
		v.name = name // safe to set it here
		pkg.AddTask(v)
	case PackageBuilder:
		pkg.AddPackage(name, v)
	case []string:
		task := NewTask(name)
		for _, s := range v {
			r := utils.MustRenderTemplate(s, pkg.reference)
			task.Add(r)
		}
		pkg.AddTask(task)
	case []Command:
		task := NewTask(name)
		for _, c := range v {
			if r, ok := c.(Renderer); ok {
				r.Render(pkg.reference)
			}
			task.Add(c)
		}
		pkg.AddTask(task)
	default:
		panic(fmt.Sprintf("type %T not supported in Package.Add", sth))
	}
}

func (pkg *packageImpl) AddPackage(name string, pkgBuilder PackageBuilder) {
	e := validatePackage(pkgBuilder)
	if e != nil {
		panic(e)
	}
	if pkg.reference != nil {
		name = utils.MustRenderTemplate(name, pkg.reference)
	}
	pkg.validateTaskName(name)
	child := &packageImpl{cacheKeyPrefix: name, reference: pkgBuilder}
	pkgBuilder.BuildPackage(child)
	for _, task := range child.Tasks() {
		pkg.AddTask(task)
	}
}

func (pkg *packageImpl) AddTask(task Task) {
	pkg.validateTaskName(task.CacheKey())
	pkg.taskNames[task.CacheKey()] = struct{}{}
	pkg.tasks = append(pkg.tasks, task)
}

func (pkg *packageImpl) precompile() (e error) {
	for _, task := range pkg.tasks {
		c, e := task.Commands()
		if e != nil {
			return e
		}
		if len(c) > 0 {
			return fmt.Errorf("pkg %q seems to be packaged already", task.CacheKey())
		}

		if tc, ok := task.(interface {
			Compile() error
		}); ok {
			e := tc.Compile()
			if e != nil {
				return e
			}
		}
	}

	return nil
}

func (pkg *packageImpl) validateTaskName(name string) {
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
