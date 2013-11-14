package host

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHostCreation(t *testing.T) {
	h, e := NewHost(HOST_TYPE_SSH)
	assert.NotNil(t, h)
	assert.Nil(t, e)

	h, e = NewHost(HOST_TYPE_DOCKER)
	assert.NotNil(t, h)
	assert.Nil(t, e)

	h, e = NewHost(-1)
	assert.Nil(t, h)
	assert.Contains(t, e.Error(), "host type must be of the HOST_TYPE_{DOCKER,SSH} const")

	h, e = NewHost(24)
	assert.Nil(t, h)
	assert.Contains(t, e.Error(), "host type must be of the HOST_TYPE_{DOCKER,SSH} const")
}

func TestHostTypePredicates(t *testing.T) {
	h, _ := NewHost(HOST_TYPE_SSH)

	assert.True(t, h.IsSshHost())
	assert.False(t, h.IsDockerHost())

	h, _ = NewHost(HOST_TYPE_DOCKER)

	assert.True(t, h.IsDockerHost())
	assert.False(t, h.IsSshHost())
}

func TestIPAddressHandling(t *testing.T) {
	h, e := NewHost(HOST_TYPE_SSH)

	assert.Equal(t, h.GetPublicIPAddress(), "")
	assert.Equal(t, h.GetVpnIPAddress(), "")

	e = h.SetPublicIPAddress("not an IP address")
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): not an IP address")

	e = h.SetPublicIPAddress("666.666.666.666")
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): 666.666.666.666")

	e = h.SetPublicIPAddress("127.0.0.1")
	assert.Nil(t, e)
	assert.Equal(t, h.GetPublicIPAddress(), "127.0.0.1")

	e = h.SetVpnIPAddress("not an IP address")
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): not an IP address")

	e = h.SetVpnIPAddress("666.666.666.666")
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): 666.666.666.666")

	e = h.SetVpnIPAddress("127.0.0.1")
	assert.Nil(t, e)
	assert.Equal(t, h.GetVpnIPAddress(), "127.0.0.1")
}

func TestUserHandling(t *testing.T) {
	h, _ := NewHost(HOST_TYPE_SSH)

	assert.Equal(t, h.GetUser(), "root")
	assert.Equal(t, h.IsSudoRequired(), false)

	h.SetUser("root")
	assert.Equal(t, h.GetUser(), "root")
	assert.Equal(t, h.IsSudoRequired(), false)

	h.SetUser("gfrey")
	assert.Equal(t, h.GetUser(), "gfrey")
	assert.Equal(t, h.IsSudoRequired(), true)
}
