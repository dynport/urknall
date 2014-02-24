// Go subpackage with urknall's command infrastructure. There are some predefined commands (types that implement the
// "Command" interface), but you can write custom commands, of course. Most commands come with helper functions to allow
// for easy construction when filling runlists.
//
// Its important to understand that most commands are just plain shell commands, that are executed on the host to be
// provisioned. The exception of the rule are commands only required for docker (like the DockerInitCommand for
// example).
package main

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
			panic(fmt.Errorf("failed rendering template: %s (%s)", e.Error(), tmplString))
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
