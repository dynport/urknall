package host

import (
	"fmt"
	"net"
)

const (
	HOST_TYPE_DOCKER = iota
	HOST_TYPE_SSH
)

type Host struct {
	hostType int    // What executor should be used (SSH or Docker)?
	publicIP net.IP // Host's IP address used to provision the system.
	vpnIP    net.IP // Host's private IP address.
	user     string // User used to log in.
}

func NewHost(hostType int) (host *Host, e error) {
	if hostType != HOST_TYPE_SSH && hostType != HOST_TYPE_DOCKER {
		return nil, fmt.Errorf("host type must be of the HOST_TYPE_{DOCKER,SSH} const")
	}
	return &Host{hostType: hostType}, nil
}

func (h *Host) IsDockerHost() bool {
	return h.hostType == HOST_TYPE_DOCKER
}

func (h *Host) IsSshHost() bool {
	return h.hostType == HOST_TYPE_SSH
}

func (h *Host) SetPublicIPAddress(ip string) (e error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("not a valid IP address (either IPv4 or IPv6): %s", ip)
	}
	h.publicIP = parsedIP
	return nil
}

func (h *Host) GetPublicIPAddress() string {
	if h.publicIP == nil {
		return ""
	}
	return h.publicIP.String()
}

func (h *Host) SetVpnIPAddress(ip string) (e error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("not a valid IP address (either IPv4 or IPv6): %s", ip)
	}
	h.vpnIP = parsedIP
	return nil
}

func (h *Host) GetVpnIPAddress() string {
	if h.vpnIP == nil {
		return ""
	}
	return h.vpnIP.String()
}

func (h *Host) SetUser(user string) {
	h.user = user
}

func (h *Host) GetUser() string {
	if h.user == "" {
		return "root"
	}
	return h.user
}

func (h *Host) IsSudoRequired() bool {
	if h.user != "" && h.user != "root" {
		return true
	}
	return false
}
