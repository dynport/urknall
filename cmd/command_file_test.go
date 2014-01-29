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

func TestAddFileForLogging(t *testing.T) {
	Convey("Given a basic file action with content being a single line", t, func() {
		fAct := &FileCommand{
			Content: "something",
			Path:    "/tmp/foo",
		}

		Convey("Then the logging line should contain information on path", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] /tmp/foo")
		})
	})

	Convey("Given a file action with different owner", t, func() {
		fAct := FileCommand{
			Content: "something",
			Path:    "/tmp/foo",
			Owner:   "gfrey",
		}

		Convey("Then the logging line should contain information on path and owner", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ][CHOWN:gfrey] /tmp/foo")
		})
	})

	Convey("Given a file action with permissions", t, func() {
		fAct := FileCommand{
			Content:     "something",
			Path:        "/tmp/foo",
			Permissions: 0644,
		}

		Convey("Then the logging line should contain information on path and permissions", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ][CHMOD:0644] /tmp/foo")
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
				e := fAct.Validate()
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, "no path given")
			})
		})

		Convey("When a path is set", func() {
			fAct.Path = "/tmp/foo"
			Convey("When no content is given", func() {
				Convey("Then the creation of the actual shell command must fail", func() {
					e := fAct.Validate()
					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, `no content given for file "/tmp/foo"`)
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
