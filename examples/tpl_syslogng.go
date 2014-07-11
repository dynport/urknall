package main

import "github.com/dynport/urknall"

const syslogNgRestart = "{ status syslog-ng | grep running && restart syslog-ng; } || start syslog-ng"

type SyslogNg struct {
	Version string `urknall:"required=true"`
}

func (ng *SyslogNg) Url() string {
	return "http://www.balabit.com/downloads/files/syslog-ng/open-source-edition/{{ .Version }}/source/syslog-ng_{{ .Version }}.tar.gz"
}

func (ng *SyslogNg) Render(r urknall.Package) {
	r.AddCommands("packages",
		InstallPackages("build-essential", "libevtlog-dev", "pkg-config", "libglib2.0-dev"),
	)
	r.AddCommands("download", DownloadAndExtract("{{ .Url }}", "/opt/src"))
	r.AddCommands("build",
		And(
			"cd {{ .SrcDir }}",
			"./configure",
			"make",
			"make install",
		),
	)
	r.AddCommands("upstart", WriteFile("/etc/init/syslog-ng.conf", syslogNgUpstart, "root", 0644))
}

func (ng *SyslogNg) SrcDir() string {
	return "/opt/src/syslog-ng-{{ .Version }}"
}

const syslogNgUpstart = `# syslog-ng - system logging daemon
#
# syslog-ng is an replacement for the traditionala syslog daemon, logging messages from applications

description     "system logging daemon"

start on filesystem
stop on runlevel [06]

env LD_LIBRARY_PATH=/usr/local/lib

respawn

exec syslog-ng -F
`

type SyslogNgReceiver struct {
	Version  string `urknall:"default=3.5.1"`
	LogsRoot string `urknall:"default=/var/log/hourly"`
	AmqpHost string
}

func (p *SyslogNgReceiver) Render(r urknall.Package) {
	r.AddTemplate("base", &SyslogNg{Version: p.Version})
	r.AddTemplate("symlink", &CreateHourlySymlinks{Root: p.LogsRoot})
	r.AddCommands("config",
		WriteFile("/usr/local/etc/syslog-ng.conf", syslogReceiver, "root", 0644),
		Shell(syslogNgRestart),
	)
}

type SyslogNgSender struct {
	Receiver string
	Version  string `urknall:"default=3.5"`
}

func (s *SyslogNgSender) Render(r urknall.Package) {
	r.AddTemplate("base",
		&SyslogNg{Version: s.Version},
	)
	r.AddCommands("config",
		WriteFile("/usr/local/etc/syslog-ng.conf", syslogNgSender, "root", 0644),
		Shell(syslogNgRestart),
	)
}

const syslogNgSender = `@version: {{ .Version }}
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

type CreateHourlySymlinks struct {
	Root string `urknall:"default=/var/log/hourly"`
}

func (*CreateHourlySymlinks) Render(r urknall.Package) {
	r.AddCommands("base",
		Mkdir("/opt/scripts", "root", 0755),
		WriteFile("/opt/scripts/create_hourly_symlinks.sh", createHourlySymlinks, "root", 0755),
		WriteFile("/etc/cron.d/create_hourly_symlinks", "* * * * * root /opt/scripts/create_hourly_symlinks.sh 2>&1 | logger -i -t create_hourly_symlinks\n", "root", 0644),
	)
}

const createHourlySymlinks = `
#!/usr/bin/env bash
set -e

LOG_DIR={{ .Root }}
NOW=$LOG_DIR/$(date +"%Y/%m/%d/%Y-%m-%dT%H.log")
TODAY=$(dirname $NOW)

mkdir -p $TODAY
touch $NOW
chmod 0644 $NOW
ln -nfs $NOW $LOG_DIR/current
ln -nfs $TODAY $LOG_DIR/today
`

const syslogReceiver = `@version: {{ .Version }}
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
