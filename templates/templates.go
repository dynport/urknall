// Templating mechanisms used in zwo.
package templates

import (
	"bytes"
	"github.com/dynport/zwo/assets"
	"text/template"
)

// Retrieve the given asset from the assets folder (file with given name must exist and be compiled using goassets from
// http://github.com/dynport/dpkg) and render it using the text/template and the information from the provided
// interface.
func RenderAssetFromString(name string, iface interface{}) (rendered string, e error) {
	asset, e := assets.Get(name)
	if e != nil {
		return "", e
	}
	return RenderTemplateFromString(string(asset), iface)
}

// Render the template from the given string using text/template and the information from the interface provided.
func RenderTemplateFromString(tmplString string, i interface{}) (rendered string, e error) {
	tpl := template.New("")
	tpl, e = tpl.Parse(tmplString)
	if e != nil {
		return "", e
	}

	resultBuffer := &bytes.Buffer{}
	e = tpl.Execute(resultBuffer, i)
	return string(resultBuffer.Bytes()), e
}
