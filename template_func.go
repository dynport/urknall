package urknall

type TemplateFunc func(Package)

func (f TemplateFunc) Render(p Package) {
	f(p)
}
