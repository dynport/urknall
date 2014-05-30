package urknall

import (
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/target"
)

// The target interface is used to describe something a package can be built
// on.
type Target interface {
	Command(cmd string) (cmd.ExecCommand, error)
	User() string
	String() string
	Reset() error
}

// Create an SSH target. The address is an identifier of the form
// `[<user>@?]<host>[:port]`. It is assumed that authentication via public key
// will work, i.e. the remote host has the building user's public key in its
// authorized_keys file.
func NewSshTarget(address string) (Target, error) {
	return target.NewSshTarget(address)
}

// Special SSH target that uses the given password for accessing the machine.
// This is required mostly for testing and shouldn't be used in production
// settings.
func NewSshTargetWithPassword(address, password string) (Target, error) {
	target, e := target.NewSshTarget(address)
	if e == nil {
		target.Password = password
	}
	return target, e
}

// Use the local host for building.
func NewLocalTarget() (Target, error) {
	return target.NewLocalTarget(), nil
}
