package urknall

// This is a short-cut usable for templates without configuration, i.e. where
// no separate struct is required.
type TemplateFunc func(Package)

// This is just required to make the short cut work.
func (f TemplateFunc) Render(p Package) {
	f(p)
}
