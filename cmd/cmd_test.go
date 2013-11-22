package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestInstallPackagesCommand(t *testing.T) {
	Convey("When the InstallPackages command is called without any packages given", t, func() {
		Convey("Then the function panics", func() {
			f := func() { InstallPackages() }
			So(f, ShouldPanicWith, "empty package list given")
		})
	})

	Convey("When the InstallPackages command is called for a package foo", t, func() {
		c := InstallPackages("foo")
		Convey("Then the result should contain the foo package", func() {
			So(c, ShouldContainSubstring, "foo")
		})
	})

	Convey("When the InstallPackages command is called for packages foo and bar", t, func() {
		c := InstallPackages("foo", "bar")
		Convey("Then the result should contain the both packages", func() {
			So(c, ShouldContainSubstring, "foo")
			So(c, ShouldContainSubstring, "bar")
		})
	})
}

func TestAndCommand(t *testing.T) {
	Convey("When the And command is called without any subcommands given", t, func() {
		Convey("Then the function panics", func() {
			f := func() { And() }
			So(f, ShouldPanicWith, "empty list of commands given")
		})
	})

	Convey("When the And command is called for a command foo", t, func() {
		c := And("foo")
		Convey("Then the result should only contain the foo command", func() {
			So(c, ShouldEqual, "foo")
		})
	})

	Convey("When the And command is called for commands foo and bar", t, func() {
		c := And("foo", "bar")
		Convey("Then the result should contain the combined commands", func() {
			So(c, ShouldEqual, "{ foo && bar; }")
		})
	})
}

func TestOrCommand(t *testing.T) {
	Convey("When the Or command is called without any subcommands given", t, func() {
		Convey("Then the function panics", func() {
			f := func() { Or() }
			So(f, ShouldPanicWith, "empty list of commands given")
		})
	})

	Convey("When the Or command is called for a command foo", t, func() {
		c := Or("foo")
		Convey("Then the result should only contain the foo command", func() {
			So(c, ShouldEqual, "foo")
		})
	})

	Convey("When the Or command is called for commands foo and bar", t, func() {
		c := Or("foo", "bar")
		Convey("Then the result should contain the combined commands", func() {
			So(c, ShouldEqual, "{ foo || bar; }")
		})
	})
}

func TestMkdirCommand(t *testing.T) {
	Convey("When the Mkdir command is called without a path", t, func() {
		Convey("Then the function panics", func() {
			f := func() { Mkdir("", "", 0) }
			So(f, ShouldPanicWith, "empty path given to mkdir")
		})
	})

	Convey("Given the path '/tmp/foo'", t, func() {
		path := "/tmp/foo"
		Convey("When neither owner nor mode are set", func() {
			owner := ""
			var mode os.FileMode = 0
			Convey("Then the mkdir command won't set owner or permissions", func() {
				c := Mkdir(path, owner, mode)
				So(c, ShouldEqual, "mkdir -p /tmp/foo")
			})
		})

		Convey("When the owner is set", func() {
			owner := "gfrey"
			var mode os.FileMode = 0
			Convey("Then the mkdir command will change the owner", func() {
				c := Mkdir(path, owner, mode)
				So(c, ShouldContainSubstring, "chown gfrey /tmp/foo")
			})
		})

		Convey("When the mode is set", func() {
			owner := ""
			var mode os.FileMode = 0755
			Convey("Then the mkdir command will change the permissions", func() {
				c := Mkdir(path, owner, mode)
				So(c, ShouldContainSubstring, "chmod 755 /tmp/foo")
			})
		})

		Convey("When both owner and mode are set", func() {
			owner := "gfrey"
			var mode os.FileMode = 0755
			Convey("Then the mkdir command will change owner and permissions", func() {
				c := Mkdir(path, owner, mode)
				So(c, ShouldContainSubstring, "chown gfrey /tmp/foo")
				So(c, ShouldContainSubstring, "chmod 755 /tmp/foo")
			})
		})
	})

}
