package cmd

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

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
				So(f, ShouldPanicWith, "no destination given")
			})
		})
	})

	Convey("Given an url and an destination", t, func() {
		url := "http://example.com/foobar.gz"
		destination := "/tmp"
		Convey("When the DownloadToFile method is called", func() {
			c := DownloadToFile(url, destination, "", 0)
			Convey("Then the result must contain a download command", func() {
				So(c.Shell(), ShouldContainSubstring, fmt.Sprintf(`curl -SsfLO "%s"`, url))
			})
			Convey("Then the result must contain a move command", func() {
				So(c.Shell(), ShouldContainSubstring, "mv /tmp/downloads/foobar.gz "+destination)
			})
		})
		Convey("Given an owner different from root", func() {
			owner := "gfrey"
			Convey("When the DownloadToFile method is called", func() {
				c := DownloadToFile(url, destination, owner, 0)
				Convey("Then the result must contain a chown command", func() {
					So(c.Shell(), ShouldContainSubstring, "chown gfrey /tmp/foobar.gz")
				})
			})
		})
		Convey("Given an file permission different from 0", func() {
			permissions := os.FileMode(0644)
			Convey("When the DownloadToFile method is called", func() {
				c := DownloadToFile(url, destination, "", permissions)
				Convey("Then the result must contain a chmod command", func() {
					So(c.Shell(), ShouldContainSubstring, "chmod 644 /tmp/foobar.gz")
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
				So(f, ShouldPanicWith, "no destination given")
			})
		})
	})

	Convey("Given an URL and a target directory", t, func() {
		url := "http://example.com/foobar.tgz"
		targetDir := "/tmp/foobar"
		Convey("When the DownloadAndExtract method is called", func() {
			c := DownloadAndExtract(url, targetDir)
			Convey("Then the result must contain a download command", func() {
				So(c.Shell(), ShouldContainSubstring, fmt.Sprintf(`curl -SsfLO "%s"`, url))
			})
			Convey("Then the result must contain an extract command", func() {
				So(c.Shell(), ShouldContainSubstring, "tar xfz /tmp/downloads/foobar.tgz")
			})
		})
	})
}
