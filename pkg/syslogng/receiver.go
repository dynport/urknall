package syslogng

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pkg/syslog"
)

type Receiver struct {
	Version  string `urknall:"default=3.5.1"`
	LogsRoot string `urknall:"default=/var/log/hourly"`
	AmqpHost string
}

func (p *Receiver) Package(r *urknall.Package) {
	r.Add(
		&Package{Version: p.Version},
		cmd.WriteFile("/usr/local/etc/syslog-ng.conf", receiver, "root", 0644),
		&syslog.CreateHourlySymlinks{Root: p.LogsRoot},
		restartCommand,
	)
}

const receiver = `@version: {{ .Version }}
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

{{ with .AmqpHost }}
destination d_amqp {
  amqp(
      vhost("/")
      host("{{ . }}")
      port(5672)
      username("guest") # required option, no default
      password("guest") # required option, no default
      exchange("syslog")
      exchange_declare(yes)
      exchange_type("fanout")
      routing_key("$HOST.$PROGRAM.$PRIORITY")
      body("$S_ISODATE $HOST $PROGRAM.$PRIORITY[$PID]: $MSG\n")
      persistent(yes)
      frac_digits(6)
      value-pairs(
          scope("selected-macros" "nv-pairs" "sdata")
      )
  );
};
{{ end }}

destination d_file {
  file(
    "{{ .LogsRoot }}/$R_YEAR/$R_MONTH/$R_DAY/$R_YEAR-$R_MONTH-${R_DAY}T${R_HOUR}.log"
    template("$S_ISODATE $HOST $PROGRAM.$PRIORITY[$PID]: $MSG\n")
    template_escape(no)
    perm( 0644 )
    dir_perm( 0775 )
    frac_digits(6)
  );
};

log {
  source(s_local);
  source(s_network);
  {{ with .AmqpHost }}destination(d_amqp);{{ end }}
  destination(d_file);
};
`
