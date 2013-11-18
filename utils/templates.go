package utils

import (
	"bytes"
	"fmt"
	"text/template"
)

// Delegates action to RenderTemplate. Panics in case of an error returned.
func MustRenderTemplate(tmplString string, i interface{}) (rendered string) {
	renderedCommand, e := RenderTemplate(tmplString, i)
	if e != nil {
		panic(fmt.Errorf("failed rendering template: %s", e.Error()))
	}
	return renderedCommand
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
