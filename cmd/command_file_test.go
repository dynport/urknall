package cmd

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestWriteFileConvenienceFunction(t *testing.T) {
	Convey("When the File function is called", t, func() {
		fc := WriteFile("some/path", "some content", "owner", 0644)
		Convey("Then a file command is created with the values set accordingly", func() {
			So(fc, ShouldNotBeNil)
			So(fc.Path, ShouldEqual, "some/path")
			So(fc.Content, ShouldEqual, "some content")
			So(fc.Owner, ShouldEqual, "owner")
			So(fc.Permissions, ShouldEqual, 0644)
		})
	})
}

func TestWriteAssetConvenienceFunction(t *testing.T) {
	Convey("When the Asset function is called with an existing asset", t, func() {
		fc := WriteAsset("some/path", "fw_upstart.sh", "owner", 0644)
		Convey("Then a file command is created with the values set accordingly", func() {
			So(fc, ShouldNotBeNil)
			So(fc.Path, ShouldEqual, "some/path")
			So(fc.Content, ShouldStartWith, "#!/bin/sh")
			So(fc.Owner, ShouldEqual, "owner")
			So(fc.Permissions, ShouldEqual, 0644)
		})
	})

	Convey("When the Asset function is called with an unknown asset", t, func() {
		f := func() { WriteAsset("some/path", "does.not.exist", "owner", 644) }
		Convey("Then the function must panic", func() {
			So(f, ShouldPanic)
		})
	})
}

func TestAddFileForLogging(t *testing.T) {
	Convey("Given a basic file action with content being a single line", t, func() {
		fAct := &FileCommand{
			Content: "something",
			Path:    "/tmp/foo",
		}

		Convey("Then the logging line should contain information on path and content", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] /tmp/foo << something")
		})
	})

	Convey("Given a file action with different owner", t, func() {
		fAct := FileCommand{
			Content: "something",
			Path:    "/tmp/foo",
			Owner:   "gfrey",
		}

		Convey("Then the logging line should contain information on path and content", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ][CHOWN:gfrey] /tmp/foo << something")
		})
	})

	Convey("Given a file action with permissions", t, func() {
		fAct := FileCommand{
			Content:     "something",
			Path:        "/tmp/foo",
			Permissions: 0644,
		}

		Convey("Then the logging line should contain information on path and content", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ][CHMOD:0644] /tmp/foo << something")
		})
	})

	Convey("Given a file action with long content", t, func() {
		fAct := FileCommand{
			Content: "123456789.123456789.123456789.123456789.123456789.123456789.",
			Path:    "/tmp/foo",
		}

		Convey("Then the logging line should truncate the content to 50 characters", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] /tmp/foo << 123456789.123456789.123456789.123456789.123456789.")
		})
	})

}

func TestAddFileForDocker(t *testing.T) {
	Convey("Given a basic file action with content being a single line", t, func() {
		fAct := FileCommand{
			Content: "something",
			Path:    "/tmp/foo",
		}

		Convey("Then the docker runfile should contain a simple command", func() {
			v := fAct.Docker()
			So(v, ShouldStartWith, "RUN")
		})
	})
}

func TestAddFileForShell(t *testing.T) {
	rawContent := "something"
	zippedContent := `H4sIAAAJbogA/yrOz00tycjMSwcAAAD//wEAAP//+zHaCQkAAAA=`
	hash := "3fc9b689459d738f8c88a3a48aa9e33542016b7a4052e001aaa536fca74813cb"
	tmpFile := fmt.Sprintf("/tmp/wunderscale.%s", hash)

	commandBase := fmt.Sprintf("mkdir -p /tmp && echo %s | base64 -d | gunzip > %s", zippedContent, tmpFile)

	Convey("Given a basic file action", t, func() {
		fAct := FileCommand{}

		Convey("When no path is set", func() {
			Convey("Then the creation of the actual shell command must fail", func() {
				So(func() { fAct.Shell() }, ShouldPanicWith, "no path given")
			})
		})

		Convey("When a path is set", func() {
			fAct.Path = "/tmp/foo"
			Convey("When no content is given", func() {
				Convey("Then the creation of the actual shell command must fail", func() {
					So(func() { fAct.Shell() }, ShouldPanicWith, "no content given")
				})
			})

			Convey("When content is given", func() {
				fAct.Content = rawContent
				Convey("Then creation of the shell command succeeds", func() {
					sc := fAct.Shell()
					So(sc, ShouldContainSubstring, commandBase)
				})

				Convey("When a owner different then root is given", func() {
					fAct.Owner = "gfrey"
					Convey("Then the shell command contains a chown call", func() {
						sc := fAct.Shell()
						So(sc, ShouldContainSubstring, "chown gfrey /tmp/wunderscale.")
					})
				})

				Convey("When a file mode other than 0 is given", func() {
					fAct.Permissions = 0644
					Convey("Then the shell command contains a chmod call", func() {
						sc := fAct.Shell()
						So(sc, ShouldContainSubstring, "chmod 644 /tmp/wunderscale.")
					})
				})
			})
		})
	})
}
