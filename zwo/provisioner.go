package zwo

import (
	"crypto/sha256"
	"fmt"
	"github.com/dynport/gossh"
	"github.com/dynport/zwo/host"
	"strings"
)

type Provisioner interface {
	Provision(packages ...Compiler) (e error)
}

type sshClient struct {
	client *gossh.Client
	host   *host.Host
}

func NewProvisioner(h *host.Host) (p Provisioner) {
	switch {
	case h.IsSshHost():
		sc := gossh.New(h.GetPublicIPAddress(), h.GetUser())
		return &sshClient{client: sc, host: h}
	case h.IsDockerHost():
		return nil
	}
	return nil
}

func (sc *sshClient) Provision(packages ...Compiler) (e error) {
	fmt.Println("starting to provision")
	if packages == nil || len(packages) == 0 {
		return fmt.Errorf("compilables must be given")
	}

	for _, pkg := range packages {
		rl := &Runlist{host: sc.host}
		rl.setConfig(pkg)
		rl.setName(getPackageName(pkg))
		if e = pkg.Compile(rl); e != nil {
			return fmt.Errorf("failed to compile: %s", e.Error())
		}
		if e = sc.provision(rl); e != nil {
			return fmt.Errorf("failed to provision: %s", e.Error())
		}
	}

	return nil
}

func getPackageName(pkg Compiler) (name string) {
	parts := strings.Split(fmt.Sprintf("%T", pkg), ".")
	last := parts[len(parts)-1]
	return strings.ToLower(last)
}

type taskData struct {
	action   action // The command to be executed.
	checksum string // The checksum of the command.
}

func (sc *sshClient) provision(rl *Runlist) (e error) {
	tasks := buildTasksForRunlist(rl)

	checksumDir := fmt.Sprintf("/var/cache/zwo/tree/%s", rl.getName())

	checksumHash, e := sc.buildChecksumHash(checksumDir)
	if e != nil {
		return fmt.Errorf("failed to build checksum hash: %s", e.Error())
	}

	for i := range tasks {
		task := tasks[i]
		if _, found := checksumHash[task.checksum]; found { // Task is cached.
			fmt.Printf("[CACHED] %s\n", task.action.Logging())
			delete(checksumHash, task.checksum) // Delete checksums of cached tasks from hash.
			continue
		}
		fmt.Printf("[EXEC  ] %s\n", task.action.Logging())
		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = sc.cleanUpRemainingCachedEntries(checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}
		rsp, e := sc.client.Execute(task.action.Shell())
		if e != nil {
			return fmt.Errorf("failed to execute cmd: '%s'\nStdErr:\n%sStdOut:\n%s", e.Error(), rsp.Stderr(), rsp.Stdout())
		}
		if _, e = sc.executeCommand(fmt.Sprintf("touch %s/%s", checksumDir, task.checksum)); e != nil {
			return e
		}
	}

	return nil
}

func buildTasksForRunlist(rl *Runlist) (tasks []*taskData) {
	tasks = make([]*taskData, 0, len(rl.actions))

	cmdHash := sha256.New()
	for i := range rl.actions {
		rawCmd := rl.actions[i].Shell()
		cmdHash.Write([]byte(rawCmd))

		task := &taskData{action: rl.actions[i], checksum: fmt.Sprintf("%x", cmdHash.Sum(nil))}
		tasks = append(tasks, task)
	}
	return tasks
}

func (sc *sshClient) executeCommand(cmdRaw string) (result *gossh.Result, e error) {
	c, e := Execute(cmdRaw)(sc.host, struct{}{})
	if e != nil {
		return nil, e
	}
	result, e = sc.client.Execute(c.Shell())
	if e != nil {
		stderr := ""
		if result != nil {
			stderr = "\n" + result.Stderr()
		}
		return result, fmt.Errorf("%s%s", e.Error(), stderr)
	}
	return result, nil
}

func (sc *sshClient) buildChecksumHash(checksumDir string) (checksumMap map[string]struct{}, e error) {
	// Make sure the directory exists.
	if _, e = sc.executeCommand(fmt.Sprintf("mkdir -p %s", checksumDir)); e != nil {
		return nil, fmt.Errorf("failed to create checksum directory: %s", e.Error())
	}

	rsp, e := sc.executeCommand(fmt.Sprintf("ls %s | xargs", checksumDir))
	if e != nil {
		return nil, e
	}
	checksums := strings.Fields(strings.TrimSpace(rsp.Stdout()))

	checksumMap = make(map[string]struct{})
	for i := range checksums {
		if len(checksums[i]) != 64 {
			return nil, fmt.Errorf("invalid checksum '%s' found in '%s'", checksums[i], checksumDir)
		}
		checksumMap[checksums[i]] = struct{}{}
	}
	return checksumMap, nil
}

func (sc *sshClient) cleanUpRemainingCachedEntries(checksumDir string, checksumHash map[string]struct{}) (e error) {
	invalidCacheEntries := make([]string, 0, len(checksumHash))
	for k, _ := range checksumHash {
		invalidCacheEntries = append(invalidCacheEntries, k)
	}
	cmd := fmt.Sprintf("cd %s && rm -f %s", checksumDir, strings.Join(invalidCacheEntries, " "))
	if _, e = sc.executeCommand(cmd); e != nil {
		return e
	}
	return nil
}
