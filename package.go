package urknall

// A template is used to modularize the urknall setting. Templates are rendered
// into a package and during rendering tasks can be added. See the Package
// description for information on how to manage tasks.
type Template interface {
	Render(pkg Package)
}

// The package is an interface. It provides the methods used to add tasks to a
// package. The packages's AddTemplate and AddCommands methods will internally
// create tasks.
//
// Nesting of templates provides a lot of flexibility as different
// configurations can be used depending on the greater context.
//
// The first argument of all three Add methods is a string. These strings are
// used as identifiers for the caching mechanism. They must be unique over all
// tasks. For nested templates the identifiers are concatenated using ".".
type Package interface {
	AddTemplate(string, Template)   // Add another template, nested below the current one.
	AddCommands(string, ...Command) // Add a new task from the given commands.
	AddTask(string, Task)           // Add the given tasks to the package with the given name.
}
