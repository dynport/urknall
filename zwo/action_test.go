package zwo

import (
	"github.com/dynport/zwo/host"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteCommand(t *testing.T) {
	h, e := host.NewHost(host.HOST_TYPE_SSH)
	assert.Nil(t, e)

	cmd, e := Execute("")(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty command given")

	cmd, e = Execute("foobar")(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "foobar")
	assert.Nil(t, e)

	cmd, e = Execute("foo{{ .Bar }}baz")(h, &struct{ Bar string }{Bar: "bar"})
	assert.Equal(t, cmd.Plain(), "foobarbaz")
	assert.Nil(t, e)

	cmd, e = Execute("foo{{ .Bur }}baz")(h, &struct{ Bar string }{Bar: "bar"})
	assert.Nil(t, cmd)
	assert.Contains(t, e.Error(), "Bur is not a field")
}

func TestInstallPackagesCommand(t *testing.T) {
	h, e := host.NewHost(host.HOST_TYPE_SSH)
	assert.Nil(t, e)

	cmd, e := InstallPackages()(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty package list given")

	cmd, e = InstallPackages("foo")(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "DEBIAN_FRONTEND=noninteractive apt-get install -y foo")
	assert.Nil(t, e)

	cmd, e = InstallPackages("foo", "bar")(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "DEBIAN_FRONTEND=noninteractive apt-get install -y foo bar")
	assert.Nil(t, e)

	cmd, e = InstallPackages("foo", "{{ .Bar }}")(h, &struct{ Bar string }{Bar: "bar"})
	assert.Equal(t, cmd.Plain(), "DEBIAN_FRONTEND=noninteractive apt-get install -y foo bar")
	assert.Nil(t, e)

	cmd, e = InstallPackages("foo", "{{ .Bur }}")(h, &struct{ Bar string }{Bar: "bar"})
	assert.Nil(t, cmd)
	assert.Contains(t, e.Error(), "Bur is not a field")
}

func TestAndCommand(t *testing.T) {
	h, e := host.NewHost(host.HOST_TYPE_SSH)
	assert.Nil(t, e)

	cmd, e := And()(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty list of commands given")

	cmd, e = And(Execute("foo"))(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "foo")
	assert.Nil(t, e)

	cmd, e = And(Execute("foo"), Execute("bar"))(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "foo && bar")
	assert.Nil(t, e)

	cmd, e = And(Execute("foo {{ .Bar }}"), Execute("boz"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Equal(t, cmd.Plain(), "foo bar && boz")
	assert.Nil(t, e)

	cmd, e = And(Execute("foo {{ .Bur }}"), Execute("boz"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Nil(t, cmd)
	assert.Contains(t, e.Error(), "Bur is not a field")
}

func TestIfCommand(t *testing.T) {
	h, e := host.NewHost(host.HOST_TYPE_SSH)
	assert.Nil(t, e)

	cmd, e := If("")(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty test given")

	cmd, e = If("-d /tmp")(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty list of commands given")

	cmd, e = If("-d /tmp", Execute("foo"))(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "test -d /tmp && { foo }")
	assert.Nil(t, e)

	cmd, e = If("-d /tmp", Execute("foo"), Execute("bar"))(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "test -d /tmp && { foo && bar }")
	assert.Nil(t, e)

	cmd, e = If("-d {{ .Bar }}", Execute("foo"), Execute("bar"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Equal(t, cmd.Plain(), "test -d bar && { foo && bar }")
	assert.Nil(t, e)

	cmd, e = If("-d /tmp", Execute("foo"), Execute("{{ .Bar }}"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Equal(t, cmd.Plain(), "test -d /tmp && { foo && bar }")
	assert.Nil(t, e)

	cmd, e = If("-d {{ .Bur }}", Execute("foo"), Execute("bar"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Nil(t, cmd)
	assert.Contains(t, e.Error(), "Bur is not a field")

	cmd, e = If("-d /tmp", Execute("{{ .Bur }}"), Execute("bar"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Nil(t, cmd)
	assert.Contains(t, e.Error(), "Bur is not a field")
}

func TestIfNotCommand(t *testing.T) {
	h, e := host.NewHost(host.HOST_TYPE_SSH)
	assert.Nil(t, e)

	cmd, e := IfNot("")(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty test given")

	cmd, e = IfNot("-d /tmp")(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty list of commands given")

	cmd, e = IfNot("-d /tmp", Execute("foo"))(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "test -d /tmp || { foo }")
	assert.Nil(t, e)

	cmd, e = IfNot("-d /tmp", Execute("foo"), Execute("bar"))(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), "test -d /tmp || { foo && bar }")
	assert.Nil(t, e)

	cmd, e = IfNot("-d {{ .Bar }}", Execute("foo"), Execute("bar"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Equal(t, cmd.Plain(), "test -d bar || { foo && bar }")
	assert.Nil(t, e)

	cmd, e = IfNot("-d /tmp", Execute("foo"), Execute("{{ .Bar }}"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Equal(t, cmd.Plain(), "test -d /tmp || { foo && bar }")
	assert.Nil(t, e)

	cmd, e = IfNot("-d {{ .Bur }}", Execute("foo"), Execute("bar"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Nil(t, cmd)
	assert.Contains(t, e.Error(), "Bur is not a field")

	cmd, e = IfNot("-d /tmp", Execute("{{ .Bur }}"), Execute("bar"))(h, &struct{ Bar string }{Bar: "bar"})
	assert.Nil(t, cmd)
	assert.Contains(t, e.Error(), "Bur is not a field")
}
