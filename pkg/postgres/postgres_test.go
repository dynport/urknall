package postgres

import (
	"github.com/dynport/urknall"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var dockerIp = os.Getenv("DOCKER_IP")

func Test(t *testing.T) {
	if dockerIp == "" {
		t.Skip("no docker host provided")
	}
	Convey("Postgres Package", t, func() {
		l, e := urknall.OpenStdoutLogger()
		if e != nil {
			t.Fatal(e.Error())
		}
		defer l.Close()
		host := &urknall.Host{IP: dockerIp}
		host.Docker = &urknall.DockerSettings{}
		pkg := &Package{}
		imageId, e := host.CreateDockerImage("ubuntu", "postgres", pkg)
		So(imageId, ShouldEqual, "test")
		So(e, ShouldBeNil)
	})
}
