#!/bin/sh

case "$IFACE" in
	{{ .PrimaryInterface }})
		/sbin/iptables-restore < /etc/iptables/rules_ipv4
		/sbin/ip6tables-restore < /etc/iptables/rules_ipv6
		;;
esac

