package syslog

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type CreateHourlySymlinks struct {
	Root string `urknall:"default=/var/log/hourly"`
}

func (*CreateHourlySymlinks) Package(r *urknall.Runlist) {
	r.Add(
		cmd.Mkdir("/opt/scripts", "root", 0755),
		cmd.WriteFile("/opt/scripts/create_hourly_symlinks.sh", createHourlySymlinks, "root", 0755),
		cmd.WriteFile("/etc/cron.d/create_hourly_symlinks", "* * * * * root /opt/scripts/create_hourly_symlinks.sh 2>&1 | logger -i -t create_hourly_symlinks\n", "root", 0644),
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
