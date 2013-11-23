package zwo

import (
	"fmt"
	"github.com/dynport/zwo/host"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAddFileForLogging(t *testing.T) {

	Convey("Given a basic file action with content being a single line", t, func() {
		h, _ := host.New("127.0.0.1")
		fAct := fileAction{
			host:    h,
			content: "something",
			path:    "/tmp/foo",
		}

		Convey("Then the logging line should contain information on path and content", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] /tmp/foo << something")
		})
	})

	Convey("Given a file action with sudo required", t, func() {
		h, _ := host.New("127.0.0.1")
		h.SetUser("gfrey")
		fAct := fileAction{
			host:    h,
			content: "something",
			path:    "/tmp/foo",
		}

		Convey("Then the logging line should contain information on path and content", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] [SUDO] /tmp/foo << something")
		})
	})

	Convey("Given a file action with different owner", t, func() {
		h, _ := host.New("127.0.0.1")
		fAct := fileAction{
			host:    h,
			content: "something",
			path:    "/tmp/foo",
			owner:   "gfrey",
		}

		Convey("Then the logging line should contain information on path and content", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] [CHOWN:gfrey] /tmp/foo << something")
		})
	})

	Convey("Given a file action with permissions", t, func() {
		h, _ := host.New("127.0.0.1")
		fAct := fileAction{
			host:    h,
			content: "something",
			path:    "/tmp/foo",
			mode:    0644,
		}

		Convey("Then the logging line should contain information on path and content", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] [CHMOD:0644] /tmp/foo << something")
		})
	})

	Convey("Given a file action with long content", t, func() {
		h, _ := host.New("127.0.0.1")
		fAct := fileAction{
			host:    h,
			content: "123456789.123456789.123456789.123456789.123456789.123456789.",
			path:    "/tmp/foo",
		}

		Convey("Then the logging line should truncate the content to 50 characters", func() {
			v := fAct.Logging()
			So(v, ShouldEqual, "[FILE   ] /tmp/foo << 123456789.123456789.123456789.123456789.123456789.")
		})
	})

}

func TestAddFileForDocker(t *testing.T) {
	h, _ := host.New("127.0.0.1")

	Convey("Given a basic file action with content being a single line", t, func() {
		fAct := fileAction{
			host:    h,
			content: "something",
			path:    "/tmp/foo",
		}

		Convey("Then the docker runfile should contain a simple command", func() {
			v := fAct.Docker()
			So(v, ShouldStartWith, "RUN")
		})
	})
}

func TestAddFileForShell(t *testing.T) {
	h, _ := host.New("127.0.0.1")

	rawContent := "something"
	zippedContent := `H4sIAAAJbogA/yrOz00tycjMSwcAAAD//wEAAP//+zHaCQkAAAA=`
	hash := "3fc9b689459d738f8c88a3a48aa9e33542016b7a4052e001aaa536fca74813cb"
	tmpFile := fmt.Sprintf("/tmp/wunderscale.%s", hash)

	commandBase := fmt.Sprintf("mkdir -p /tmp && echo %s | base64 -d | gunzip > %s", zippedContent, tmpFile)

	Convey("Given a basic file action", t, func() {
		fAct := fileAction{
			host: h,
		}

		Convey("When no path is set", func() {
			Convey("Then the creation of the actual shell command must fail", func() {
				So(func() { fAct.Shell() }, ShouldPanicWith, "no path given")
			})
		})

		Convey("When a path is set", func() {
			fAct.path = "/tmp/foo"
			Convey("When no content is given", func() {
				Convey("Then the creation of the actual shell command must fail", func() {
					So(func() { fAct.Shell() }, ShouldPanicWith, "no content given")
				})
			})

			Convey("When content is given", func() {
				fAct.content = rawContent
				Convey("Then creation of the shell command succeeds", func() {
					sc := fAct.Shell()
					So(sc, ShouldContainSubstring, commandBase)
				})

				Convey("When a owner different then root is given", func() {
					fAct.owner = "gfrey"
					Convey("Then the shell command contains a chown call", func() {
						sc := fAct.Shell()
						So(sc, ShouldContainSubstring, "chown gfrey /tmp/wunderscale.")
					})
				})

				Convey("When a file mode other than 0 is given", func() {
					fAct.mode = 0644
					Convey("Then the shell command contains a chmod call", func() {
						sc := fAct.Shell()
						So(sc, ShouldContainSubstring, "chmod 644 /tmp/wunderscale.")
					})
				})
			})
		})
	})
}
