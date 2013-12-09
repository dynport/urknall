package zwo

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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

func (t *testPkgWoutCNamer) Pack(r *Runlist) {
}

type testPkgWCNamer struct{}

func (t *testPkgWCNamer) Pack(r *Runlist) {
}

func (t *testPkgWCNamer) PackageName() string {
	return "Rumpelstilzchen"
}

func TestPackageName(t *testing.T) {
	Convey("Given a package not implementing the PackageNamer interface", t, func() {
		pkg := &testPkgWoutCNamer{}
		Convey("When the package's name is retrieved", func() {
			name := packageName(pkg)
			Convey("Then it is equal to the package's internal name in lower case", func() {
				So(name, ShouldEqual, "zwo.testpkgwoutcnamer")
			})
		})
	})

	Convey("Given a package implementing the PackageNamer interface", t, func() {
		pkg := &testPkgWCNamer{}
		Convey("When the package's name is retrieved", func() {
			name := packageName(pkg)
			Convey("Then it is equal to the name returned from the PackageName method in lowercase", func() {
				So(name, ShouldEqual, "rumpelstilzchen")
			})
		})
	})
}
