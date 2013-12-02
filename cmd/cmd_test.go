package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestUpdatePackagesCommand(t *testing.T) {
	Convey("When the UpdatePackages command is called", t, func() {
		cmd := UpdatePackages()
		Convey("Then the command must contain apt-get update", func() {
			So(cmd, ShouldContainSubstring, "apt-get update")
		})
		Convey("Then the command must contain apt-get upgrade", func() {
			So(cmd, ShouldContainSubstring, "apt-get upgrade")
		})
	})
}

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

func TestIfCommand(t *testing.T) {
	Convey("When the If command is called without a test", t, func() {
		f := func() { If("", "") }
		Convey("Then the function panics", func() {
			So(f, ShouldPanicWith, "empty test given")
		})
	})

	Convey("When the If command is called with a test", t, func() {
		test := "-d /tmp"
		Convey("When the If command is called without a command", func() {
			f := func() { If(test, "") }
			Convey("Then the function panics", func() {
				So(f, ShouldPanicWith, "empty command given")
			})
		})
	})

	Convey("Given the test '-d /tmp'", t, func() {
		test := "-d /tmp"
		Convey("Given the command 'echo \"true\"'", func() {
			cmd := "echo \"true\""
			Convey("Then the resulting command will contain both", func() {
				c := If(test, cmd)
				So(c, ShouldContainSubstring, test)
				So(c, ShouldContainSubstring, cmd)
			})
		})
	})
}

func TestIfNotCommand(t *testing.T) {
	Convey("When the IfNot command is called without a test", t, func() {
		f := func() { IfNot("", "") }
		Convey("Then the function panics", func() {
			So(f, ShouldPanicWith, "empty test given")
		})
	})

	Convey("When the IfNot command is called with a test", t, func() {
		test := "-d /tmp"
		Convey("When the IfNot command is called without a command", func() {
			f := func() { IfNot(test, "") }
			Convey("Then the function panics", func() {
				So(f, ShouldPanicWith, "empty command given")
			})
		})
	})

	Convey("Given the test '-d /tmp'", t, func() {
		test := "-d /tmp"
		Convey("Given the command 'echo \"true\"'", func() {
			cmd := "echo \"true\""
			Convey("When the IfNot command is called with those", func() {
				c := IfNot(test, cmd)
				Convey("Then the result must contain both", func() {
					So(c, ShouldContainSubstring, test)
					So(c, ShouldContainSubstring, cmd)
				})
			})
		})
	})
}

func TestDownloadToFileCommand(t *testing.T) {
	Convey("Given an empty URL", t, func() {
		url := ""
		destination := ""
		Convey("When the DownloadToFile method is called", func() {
			f := func() { DownloadToFile(url, destination, "", 0) }
			Convey("Then the DownloadToFile method should panic", func() {
				So(f, ShouldPanicWith, "empty url given")
			})
		})
	})

	Convey("Given an url, but an empty destination", t, func() {
		url := "http://example.com/foobar.gz"
		destination := ""
		Convey("When the DownloadToFile method is called", func() {
			f := func() { DownloadToFile(url, destination, "", 0) }
			Convey("Then the DownloadToFile method should panic", func() {
				So(f, ShouldPanicWith, "empty destination given")
			})
		})
	})

	Convey("Given an url and an destination", t, func() {
		url := "http://example.com/foobar.gz"
		destination := "/tmp"
		Convey("When the DownloadToFile method is called", func() {
			c := DownloadToFile(url, destination, "", 0)
			Convey("Then the result must contain a download command", func() {
				So(c, ShouldContainSubstring, "curl -SsfLO "+url)
			})
			Convey("Then the result must contain a move command", func() {
				So(c, ShouldContainSubstring, "mv /tmp/downloads/foobar.gz "+destination)
			})
		})
		Convey("Given an owner different from root", func() {
			owner := "gfrey"
			Convey("When the DownloadToFile method is called", func() {
				c := DownloadToFile(url, destination, owner, 0)
				Convey("Then the result must contain a chown command", func() {
					So(c, ShouldContainSubstring, "chown gfrey /tmp/foobar.gz")
				})
			})
		})
		Convey("Given an file permission different from 0", func() {
			permissions := os.FileMode(0644)
			Convey("When the DownloadToFile method is called", func() {
				c := DownloadToFile(url, destination, "", permissions)
				Convey("Then the result must contain a chmod command", func() {
					So(c, ShouldContainSubstring, "chmod 644 /tmp/foobar.gz")
				})
			})
		})
	})
}

func TestDownloadAndExtractComamnd(t *testing.T) {
	Convey("Given an empty URL", t, func() {
		url := ""
		targetDir := ""
		Convey("When the DownloadAndExtract method is called", func() {
			f := func() { DownloadAndExtract(url, targetDir) }
			Convey("Then the DownloadAndExtract method should panic", func() {
				So(f, ShouldPanicWith, "empty url given")
			})
		})
	})

	Convey("Given an URL but an empty target directory", t, func() {
		url := "http://example.com/foobar.tgz"
		targetDir := ""
		Convey("When the DownloadAndExtract method is called", func() {
			f := func() { DownloadAndExtract(url, targetDir) }
			Convey("Then the DownloadAndExtract method should panic", func() {
				So(f, ShouldPanicWith, "empty target directory given")
			})
		})
	})

	Convey("Given an URL but an empty target directory", t, func() {
		url := "http://example.com/foobar.tgz"
		targetDir := "/tmp/foobar"
		Convey("When the DownloadAndExtract method is called", func() {
			c := DownloadAndExtract(url, targetDir)
			Convey("Then the result must contain a download command", func() {
				So(c, ShouldContainSubstring, "curl -SsfLO "+url)
			})
			Convey("Then the result must contain an extract command", func() {
				So(c, ShouldContainSubstring, "tar xvfz /tmp/downloads/foobar.tgz")
			})
		})
	})

}
