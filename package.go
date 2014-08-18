package urknall

// A template is used to modularize the urknall setting. Templates are rendered
// into a package and during rendering the tasks can be added. See the Package
// description for information on how to manage tasks.
type Template interface {
	Render(pkg Package)
}

// The package is an interface. Users won't implement it themselves though. It
// provides the methods used to add tasks to a package (AddTemplate and
// AddCommands will internally create tasks, so that in most cases users won't
// have to manage those by themselves).
//
// Nesting of templates provides a lot of flexibility as different
// configurations can be used depending on the greater context.
//
// The first argument of all three Add methods is a string. These strings are
// used as identifiers for the hashing mechanism. They must be unique over all
// tasks! For nested templates the identifiers are concatenated using ".".
type Package interface {
	AddTemplate(string, Template)   // Add another template, nested below the current one.
	AddCommands(string, ...Command) // Add a new task from the given commands.
	AddTask(string, Task)           // Add the given tasks to the package with the given name.

	Build(*Build) error
}
