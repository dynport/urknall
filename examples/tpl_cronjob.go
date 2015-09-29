package main

import (
	"os"
	"github.com/dynport/urknall"
)

type Cronjob struct {
	// Name of the cron job. Used for file names and logging.
	Name     string `urknall:"required=true"`
	// The script to be executed.
	Script   []byte `urknall:"required=true"`
	// The common cron job pattern, i.e. something like '*/5 * * * *' for every five minutes.
	Pattern  string `urknall:"required=true"`
	// File mode of the script added.
	Mode     int
}

func (job *Cronjob) Render(r urknall.Package) {
	mode := os.FileMode(job.Mode)
	if mode == 0 {
		mode = 0755
	}
	scriptPath := "/opt/cron/" + job.Name
	cronPath := "/etc/cron.d/" + job.Name

	r.AddCommands("script", WriteFile(scriptPath, string(job.Script), "root", mode))
	r.AddCommands("cron", WriteFile(cronPath, job.Pattern + " root " + scriptPath + " 2>&1 | logger -i -t " + job.Name + "\n", "root", 0644))
}