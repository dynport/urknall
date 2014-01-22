package openvpn

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

func New() *Package {
	return &Package{}
}

type Package struct {
	Country  string `urknall:"required=true"`
	Province string `urknall:"required=true"`
	City     string `urknall:"required=true"`
	Org      string `urknall:"required=true"`
	Email    string `urknall:"required=true"`
	Name     string `urknall:"required=true"`
	Address  string `urknall:"required=true"`
	Netmask  string `urknall:"required=true"`

	PublicIp string `urknall:"required=true"`

	Routes []string
}

const packagePath = "/opt/package_openvpn_key"

func (p *Package) Package(r *urknall.Runlist) {
	if len(p.Country) != 2 {
		panic("Country must be exactly 2 characters long")
	}
	if len(p.Province) != 2 {
		panic("Province must be exactly 2 characters long")
	}
	r.Add(
		cmd.InstallPackages("openvpn", "iptables", "zip"),
		"cp -R /usr/share/doc/openvpn/examples/easy-rsa/2.0 /etc/openvpn/easy-rsa/",
		cmd.WriteFile("/etc/openvpn/easy-rsa/vars", vars, "root", 0644),
		"ln -nfs /etc/openvpn/easy-rsa/openssl-1.0.0.cnf /etc/openvpn/easy-rsa/openssl.cnf",
		`bash -c "cd /etc/openvpn/easy-rsa && source ./vars && ./clean-all"`,
		`bash -c "cd /etc/openvpn/easy-rsa && source ./vars && ./pkitool --initca"`,
		`bash -c "cd /etc/openvpn/easy-rsa && source ./vars && ./pkitool --server {{ .Name }}"`,
		`bash -c "cd /etc/openvpn/easy-rsa && source ./vars && ./build-dh"`,
		`bash -c "cd /etc/openvpn/easy-rsa/keys && cp -v {{ .Name }}.{crt,key} ca.crt dh1024.pem /etc/openvpn/"`,
		cmd.WriteFile("/etc/openvpn/server.conf", server, "root", 0644),
		cmd.WriteFile(packagePath, packageKey, "root", 0755),
		"{ /etc/init.d/openvpn status && /etc/init.d/openvpn restart; } || /etc/init.d/openvpn start",
	)
}

const server = `port 1194
proto tcp
dev tun0
ca ca.crt
cert {{ .Name }}.crt
key {{ .Name }}.key
dh dh1024.pem
server {{ .Address }} {{ .Netmask }}
ifconfig-pool-persist ipp.txt
{{ range .Routes }}
push "{{ . }}"
{{ end }}
keepalive 10 120
comp-lzo
persist-key
persist-tun
status openvpn-status.log
verb 3
client-to-client
`

const packageKey = `#!/usr/bin/env bash
set -e

LOGIN=$1
KEYS_DIR=/etc/openvpn/easy-rsa/keys
LOGIN_DIR=$KEYS_DIR/$LOGIN.tblk
CONFIG_PATH=$LOGIN_DIR/$LOGIN.conf
PUBLIC_IP={{ .PublicIp }}
CRT_PATH=$KEYS_DIR/$LOGIN.crt
KEY_PATH=$KEYS_DIR/$LOGIN.key

TBLK_NAME=$LOGIN.tblk
TBLK_PATH=$KEYS_DIR/$TBLK_NAME.zip

OPENVPN_NAME=$LOGIN.openvpn.zip
OPENVPN_PATH=$KEYS_DIR/$OPENVPN_NAME

if [ ! -e $CRT_PATH ]; then
  echo "ERROR: key not generated"
  exit 1
fi

rm -Rf $LOGIN_DIR
mkdir -p $LOGIN_DIR
cp -v /etc/openvpn/ca.crt $CRT_PATH $KEY_PATH $LOGIN_DIR/

echo "client
dev tun
proto tcp
remote $PUBLIC_IP 1194
resolv-retry infinite
nobind
persist-key
persist-tun
ca ca.crt
cert $LOGIN.crt
key $LOGIN.key
ns-cert-type server
comp-lzo
verb 3" > $CONFIG_PATH

cd $KEYS_DIR
zip -r $TBLK_PATH $TBLK_NAME
echo "wrote $TBLK_PATH"

cd $KEYS_DIR/$TBLK_NAME
zip $OPENVPN_PATH *.*
echo "wrote $OPENVPN_PATH"
`

const vars = `
export EASY_RSA="$(pwd)"
export OPENSSL="openssl"
export PKCS11TOOL="pkcs11-tool"
export GREP="grep"
export KEY_CONFIG=$($EASY_RSA/whichopensslcnf $EASY_RSA)
export KEY_DIR="$EASY_RSA/keys"
export PKCS11_MODULE_PATH="dummy"
export PKCS11_PIN="dummy"
export KEY_SIZE=1024
export CA_EXPIRE=3650
export KEY_EXPIRE=3650
export KEY_COUNTRY="{{ .Country }}"
export KEY_PROVINCE="{{ .Province }}"
export KEY_CITY="{{ .City }}"
export KEY_ORG="{{ .Org }}"
export KEY_EMAIL="{{ .Email }}"
export KEY_CN=
export KEY_NAME=
export KEY_OU=
export PKCS11_MODULE_PATH=changeme
export PKCS11_PIN=1234
`
