package zwo

import (
	"strings"
)

func padToFixedLength(s string, l int) string {
	if len(s) < l {
		return strings.Repeat(" ", l-len(s)) + s
	} else {
		return s[:l]
	}
}
