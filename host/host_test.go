package host

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIPAddressHandling(t *testing.T) {
	h, e := New("")
	assert.Equal(t, e.Error(), "no IP address given")

	h, e = New("not an IP address")
	assert.Nil(t, h)
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): not an IP address")

	h, e = New("666.666.666.666")
	assert.Nil(t, h)
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): 666.666.666.666")

	h, e = New("127.0.0.1")
	assert.Nil(t, e)
	assert.Equal(t, h.GetIPAddress(), "127.0.0.1")

	h, e = New("127.0.0.2")
	assert.Nil(t, e)
	assert.Equal(t, h.GetIPAddress(), "127.0.0.2")
}

func TestUserHandling(t *testing.T) {
	h, _ := New("127.0.0.1")

	assert.Equal(t, h.GetUser(), "root")
	assert.Equal(t, h.IsSudoRequired(), false)

	h.SetUser("root")
	assert.Equal(t, h.GetUser(), "root")
	assert.Equal(t, h.IsSudoRequired(), false)

	h.SetUser("gfrey")
	assert.Equal(t, h.GetUser(), "gfrey")
	assert.Equal(t, h.IsSudoRequired(), true)
}
