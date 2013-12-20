// Util functions that don't fit anywhere else.
package utils

import (
	"bytes"
	"fmt"
	"text/template"
)

// Delegates action to RenderTemplate. Panics in case of an error returned.
func MustRenderTemplate(tmplString string, i interface{}) (rendered string) {
	for j := 0; j < 8; j++ {
		renderedCommand, e := RenderTemplate(tmplString, i)
		if e != nil {
			panic(fmt.Errorf("failed rendering template: %s", e.Error()))
		}
		if renderedCommand == tmplString {
			return renderedCommand
		}
		tmplString = renderedCommand
	}
	panic("found rendering loop. max 8 levels are allowed")
}

// Render the template from the given string using text/template and the information from the interface provided.
func RenderTemplate(tmplString string, i interface{}) (rendered string, e error) {
	tpl := template.New("")
	tpl, e = tpl.Parse(tmplString)
	if e != nil {
		return "", e
	}

	resultBuffer := &bytes.Buffer{}
	e = tpl.Execute(resultBuffer, i)
	return string(resultBuffer.Bytes()), e
}
