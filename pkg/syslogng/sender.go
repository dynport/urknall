package syslogng

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Sender struct {
	Receiver string
	Version  string `urknall:"default=3.5"`
}

func (s *Sender) Package(r *urknall.Package) {
	r.Add(
		&Package{Version: s.Version},
		cmd.WriteFile("/usr/local/etc/syslog-ng.conf", sender, "root", 0644),
		restartCommand,
	)
}

const sender = `@version: {{ .Version }}
@include "scl.conf"

options {
  chain_hostnames(0);
  keep_hostname(yes);
  time_reopen(10);
  time_reap(360);
  log_fifo_size(2048);
  create_dirs(yes);
  perm(0640);
  dir_perm(0755);
  use_dns(no);
  stats_freq(43200);
  frac_digits(6);
  ts_format(iso);
};

source s_network {
  udp(port(514));
  tcp(port(514));
};

source s_local {
    file("/proc/kmsg");
    unix-stream("/dev/log");
    internal();
};

destination d_syslog_tcp {
	syslog("{{ .Receiver }}" transport("tcp"));
};

log {
	source(s_local);
	source(s_network);
	destination(d_syslog_tcp);
};
`
