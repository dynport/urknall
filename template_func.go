package urknall

// This is a short-cut usable for templates without configuration, i.e. where
// no separate struct is required.
type TemplateFunc func(Package)

func (f TemplateFunc) Render(p Package) {
	f(p)
}
