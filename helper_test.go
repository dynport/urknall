package urknall

import (
	"fmt"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPadToFixedLength(t *testing.T) {
	type test struct {
		src    string
		length int
		dest   string
	}
	tests := []test{
		{"", 2, "  "},
		{"a", 2, " a"},
		{"ab", 2, "ab"},
		{"abc", 2, "ab"},
	}

	for _, test := range tests {
		Convey(fmt.Sprintf(`Given the string %q and the length "%d"`, test.src, test.length), t, func() {
			Convey(fmt.Sprintf(`Then padToFixedLength must result in %q`, test.dest), func() {
				So(padToFixedLength(test.src, test.length), ShouldEqual, test.dest)
			})
		})
	}
}

type testPkgWoutCNamer struct{}

func (t *testPkgWoutCNamer) Package(r *Package) {
}

type testPkgWCNamer struct{}

func (t *testPkgWCNamer) Package(r *Package) {
}

func (t *testPkgWCNamer) PackageName() string {
	return "Rumpelstilzchen"
}
