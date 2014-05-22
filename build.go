package urknall

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

func Run(target Target, pkgBuilder PackageBuilder) (e error) {
	return (Build{Target: target, PkgBuilder: pkgBuilder}).Run()
}

type Build struct {
	Target
	PkgBuilder PackageBuilder
	pkg        *packageImpl
	DryRun     bool
	Env        []string
}

func (b Build) Run() error {
	pkg, e := build(b.PkgBuilder)
	if e != nil {
		return e
	}
	return b.buildPackage(pkg)
}

func (build *Build) prepare() error {
	if build.User() == "" {
		return fmt.Errorf("User not set")
	}
	cmd, e := build.prepareCommand(fmt.Sprintf(`{ grep "^%s:" /etc/group | grep %s; } && [ -d /var/lib/urknall ]`, ukGROUP, build.User()))
	if e != nil {
		return e
	}
	if e := cmd.Run(); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [ -d %[1]s ] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, build.User()),
		}

		cmd, e = build.prepareCommand(strings.Join(cmds, " && "))
		if e != nil {
			return e
		}
		out := &bytes.Buffer{}
		err := &bytes.Buffer{}
		cmd.SetStderr(err)
		cmd.SetStdout(out)
		if e := cmd.Run(); e != nil {
			return fmt.Errorf("failed to initiate user %q for provisioning: %s, out=%q err=%q", build.User(), e, out.String(), err.String())
		}
		build.Reset()
	}
	return nil
}

func (build *Build) buildPackage(pkg Package) (e error) {
	ct, e := build.buildChecksumTree()
	if e != nil {
		return e
	}

	for _, task := range pkg.Tasks() {
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvision, Hostname: build.hostname()}
		m.Publish("started")
		if e = build.buildTask(task, ct); e != nil {
			m.PublishError(e)
			return e
		}
		m.Publish("finished")
	}
	return nil
}

func (build *Build) buildTask(task Task, ct checksumTree) (e error) {
	commands, e := task.Commands()
	if e != nil {
		return e
	}
	cacheKey := task.CacheKey()
	if cacheKey == "" {
		return fmt.Errorf("CacheKey must not be empty")
	}
	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", task.CacheKey())

	var found bool
	var checksumHash map[string]struct{}
	if checksumHash, found = ct[cacheKey]; !found {
		ct[cacheKey] = map[string]struct{}{}
		checksumHash = ct[cacheKey]

		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)

		cmd, e := build.prepareCommand(createChecksumDirCmd)
		if e != nil {
			return e
		}
		err := &bytes.Buffer{}

		cmd.SetStderr(err)

		if e := cmd.Run(); e != nil {
			return fmt.Errorf(err.String() + ": " + e.Error())
		}
	}

	for _, cmd := range commands {
		logMsg := cmd.Logging()
		checksum := commandChecksum(cmd)
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvisionTask, TaskChecksum: checksum, Message: logMsg, Hostname: build.hostname(), RunlistName: cacheKey}
		if _, found := checksumHash[checksum]; found { // Task is cached.
			m.ExecStatus = pubsub.StatusCached
			m.Publish("finished")
			delete(checksumHash, checksum) // Delete checksums of cached tasks from hash.
			continue
		}

		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = build.cleanUpRemainingCachedEntries(checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}
		m.ExecStatus = pubsub.StatusExecStart
		if build.DryRun {
			m.Publish("executed")
		} else {
			m.Publish("started")
			e = executeCommand(cmd, build, checksumDir)
			m.Error = e
			m.ExecStatus = pubsub.StatusExecFinished
			m.Publish("finished")
		}
		if e != nil {
			return e
		}
	}

	return nil
}

type checksumTree map[string]map[string]struct{}

func (build *Build) buildChecksumTree() (ct checksumTree, e error) {
	ct = checksumTree{}

	cmd, e := build.prepareCommand(fmt.Sprintf(`[ -d %[1]s ] && find %[1]s -type f -name \*.done`, ukCACHEDIR))
	if e != nil {
		return nil, e
	}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}

	cmd.SetStdout(out)
	cmd.SetStderr(err)

	if e := cmd.Run(); e != nil {
		return nil, e
	}
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		pkgname := filepath.Dir(strings.TrimPrefix(line, ukCACHEDIR+"/"))
		checksum := strings.TrimSuffix(filepath.Base(line), ".done")
		if len(checksum) != 64 {
			return nil, fmt.Errorf("invalid checksum %q found for package %q", checksum, pkgname)
		}
		if _, found := ct[pkgname]; !found {
			ct[pkgname] = map[string]struct{}{}
		}
		ct[pkgname][checksum] = struct{}{}
	}

	return ct, nil
}

func (build *Build) cleanUpRemainingCachedEntries(checksumDir string, checksumHash map[string]struct{}) (e error) {
	invalidCacheEntries := make([]string, 0, len(checksumHash))
	for k, _ := range checksumHash {
		invalidCacheEntries = append(invalidCacheEntries, fmt.Sprintf("%s.done", k))
	}
	if build.DryRun {
		(&pubsub.Message{Key: pubsub.MessageCleanupCacheEntries, InvalidatedCacheEntries: invalidCacheEntries, Hostname: build.hostname()}).Publish(".dryrun")
	} else {
		cmd := fmt.Sprintf("cd %s && rm -f *.failed %s", checksumDir, strings.Join(invalidCacheEntries, " "))
		m := &pubsub.Message{Key: pubsub.MessageUrknallInternal, Hostname: build.hostname()}
		m.Publish("started")

		c, e := build.prepareCommand(cmd)
		if e != nil {
			return e
		}
		if e := c.Run(); e != nil {
			return e
		}
		//m.sshResult = result
		m.Publish("finished")
	}
	return nil
}

func (build *Build) prepareCommand(cmd string) (cmd.ExecCommand, error) {
	var sudo string
	if build.User() != "root" {
		sudo = "sudo "
	}
	cmd = fmt.Sprintf(sudo+"sh -x -e <<EOC\n%s\nEOC\n", cmd)
	return build.Command(cmd)
}

func (build *Build) hostname() string {
	if s, ok := build.Target.(fmt.Stringer); ok {
		return s.String()
	}
	return "MISSING"
}
