package templates

import (
	"bytes"
	"github.com/dynport/zwo/assets"
	"text/template"
)

func RenderAssetFromString(name string, iface interface{}) (rendered string, e error) {
	asset, e := assets.Get(name)
	if e != nil {
		return "", e
	}
	return RenderTemplateFromString(string(asset), iface)
}

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
