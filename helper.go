package zwo

import (
	"fmt"
	"strings"
)

func padToFixedLength(s string, l int) string {
	if len(s) < l {
		return strings.Repeat(" ", l-len(s)) + s
	} else {
		return s[:l]
	}
}

func packageName(pkg Packager) (name string) {
	switch p := pkg.(type) {
	case PackageNamer:
		return strings.ToLower(p.PackageName())
	default:
		pkgName := fmt.Sprintf("%T", pkg)
		return strings.ToLower(pkgName[1:])
	}
}
