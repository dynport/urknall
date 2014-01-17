package openvpn

import (
	"github.com/dynport/urknall"
)

func (u *User) Package(r *urknall.Runlist) {
	r.Add(
		addUser,
		packagePath+" "+u.Login,
	)
}

type User struct {
	Login string `urknall:"required=true"`
	Name  string `urknall:"required=true"`
	Email string `urknall:"required=true"`
}

const addUser = `bash -xe <<EOF
cd /etc/openvpn/easy-rsa
source ./vars
export KEY_EMAIL="{{ .Email }}"
export KEY_NAME="{{ .Name }}"
/etc/openvpn/easy-rsa/pkitool {{ .Login }}
EOF
`
