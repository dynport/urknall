#!/bin/sh
set -e

case "$IFACE" in
	{{ .Interface }})
		/usr/sbin/ipset restore -! < /etc/iptables/ipsets
		/sbin/iptables-restore < /etc/iptables/rules_ipv4
		/sbin/ip6tables-restore < /etc/iptables/rules_ipv6
		;;
esac

