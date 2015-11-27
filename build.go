package urknall

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/dynport/dgtk/confirm"
	"github.com/dynport/gocli"
	"github.com/dynport/urknall/pubsub"
	"github.com/dynport/urknall/target"
)

// A shortcut creating and running a build from the given target and template.
func Run(target Target, tpl Template) (e error) {
	return (&Build{Target: target, Template: tpl}).Run()
}

// A shortcut creating and runnign a build from the given target and template
// with the DryRun flag set to true. This is quite helpful to actually see
// which commands would be exeucted in the current setting, without actually
// doing anything.
func DryRun(target Target, tpl Template) (e error) {
	return (&Build{Target: target, Template: tpl}).DryRun()
}

// A build is the glue between a target and template.
type Build struct {
	Target             // Where to run the build.
	Template           // What to actually build.
	Env       []string // Environment variables in the form `KEY=VALUE`.
	maxLength int      // length of the longest key to be executed
}

// This will render the build's template into a package and run all its tasks.
func (b *Build) Run() error {
	i, err := renderTemplate(b.Template)
	if err != nil {
		return err
	}
	m, err := readState(b.Target)
	if err != nil {
		return err
	}
	actions := confirm.Actions{}

	for _, t := range i.tasks {
		ex := []string{}
		if s, ok := m[t.name]; ok {
			ex = s.runSHAs
		}

		diff := []string{}
		checksums := []string{}
		broken := false
		for i, c := range t.commands {
			cs := c.Checksum()
			checksums = append(checksums, "/var/lib/urknall/"+t.name+"/"+cs+".done")
			if broken || len(ex) <= i || ex[i] != cs {
				diff = append(diff, cs)
				broken = true

				var pl []byte
				_, cmd, ok, err := extractWriteFile(c.command.Shell())
				if err == nil && ok {
					pl = []byte(cmd)
				}
				if len(t.name) > b.maxLength {
					b.maxLength = len(t.name)
				}
				actions.Update(t.name+" "+c.LogMsg(), pl, b.commandAction(t.name, checksums, c))
			}
		}
	}
	if err := confirm.ConfirmHTML(actions...); err != nil {
		return err
	}
	return nil
}

func (b *Build) DryRun() error {
	pkg, e := b.prepareBuild()
	if e != nil {
		return e
	}

	for _, task := range pkg.tasks {
		for _, command := range task.commands {
			m := message(pubsub.MessageTasksProvisionTask, b.hostname(), task.name)
			m.TaskChecksum = command.Checksum()
			m.Message = command.LogMsg()

			switch {
			case command.cached:
				m.ExecStatus = pubsub.StatusCached
				m.Publish("finished")
			default:
				m.ExecStatus = pubsub.StatusExecStart
				m.Publish("executed")
			}
		}
	}
	return nil
}

func (build *Build) prepareBuild() (*packageImpl, error) {
	pkg, e := renderTemplate(build.Template)
	if e != nil {
		return nil, e
	}

	if e = build.prepareTarget(); e != nil {
		return nil, e
	}

	ct, e := build.buildChecksumTree()
	if e != nil {
		return nil, fmt.Errorf("error building checksum tree: %s", e.Error())
	}

	return pkg, build.prepareTasks(ct, pkg.tasks...)
}

func (build *Build) prepareTarget() error {
	if build.User() == "" {
		return fmt.Errorf("User not set")
	}
	rawCmd := fmt.Sprintf(`{ grep "^%s:" /etc/group | grep %s; } && [ -d %[3]s ] && [ -f %[3]s/.v2 ]`,
		ukGROUP, build.User(), ukCACHEDIR)
	cmd, e := build.prepareInternalCommand(rawCmd)
	if e != nil {
		return e
	}
	if e := cmd.Run(); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [ -d %[1]s ] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, build.User()),
			fmt.Sprintf(`[ -f %[1]s/.v2 ] || { export DATE=$(date "+%%Y%%m%%d_%%H%%M%%S") && ls %[1]s | while read dir; do ls -t %[1]s/$dir/*.done | tac > %[1]s/$dir/$DATE.run; done && touch %[1]s/.v2;  } `, ukCACHEDIR),
		}

		cmd, e = build.prepareInternalCommand(strings.Join(cmds, " && "))
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
		return build.Reset()
	}
	return nil
}

func (build *Build) prepareTasks(ct checksumTree, tasks ...*task) error {
	for _, task := range tasks {
		if err := build.prepareTask(task, ct); err != nil {
			return err
		}
	}
	return nil
}

func (build *Build) prepareTask(tsk *task, ct checksumTree) (e error) {
	cacheKey := tsk.name
	if cacheKey == "" {
		return fmt.Errorf("CacheKey must not be empty")
	}
	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", tsk.name)

	var found bool
	var checksumList []string

	if checksumList, found = ct[cacheKey]; !found {
		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)

		cmd, e := build.prepareInternalCommand(createChecksumDirCmd)
		if e != nil {
			return e
		}
		err := &bytes.Buffer{}

		cmd.SetStderr(err)

		if e := cmd.Run(); e != nil {
			return fmt.Errorf(err.String() + ": " + e.Error())
		}
	}

	// find commands that need not be executed
	for i, cmd := range tsk.commands {
		checksum, e := commandChecksum(cmd.command)
		if e != nil {
			return e
		}

		switch {
		case len(checksumList) <= i || checksum != checksumList[i]:
			return nil
		default:
			cmd.cached = true
		}
	}

	return nil
}

func (build *Build) buildTask(tsk *task) (e error) {
	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", tsk.name)

	tsk.started = time.Now()

	for _, cmd := range tsk.commands {
		checksum := cmd.Checksum()

		m := message(pubsub.MessageTasksProvisionTask, build.hostname(), tsk.name)
		m.TaskChecksum = checksum
		m.Message = cmd.LogMsg()

		var cmdErr error

		switch {
		case cmd.cached:
			m.ExecStatus = pubsub.StatusCached
		default:
			m.ExecStatus = pubsub.StatusExecStart
			m.Publish("started")

			r := &commandRunner{
				build:    build,
				command:  cmd.command,
				dir:      checksumDir,
				taskName: tsk.name,
			}
			cmdErr = r.run()

			m.Error = cmdErr
			m.ExecStatus = pubsub.StatusExecFinished
		}
		m.Publish("finished")
		err := build.addCmdToTaskLog(tsk, checksumDir, checksum, cmdErr)
		switch {
		case cmdErr != nil:
			return cmdErr
		case err != nil:
			return err
		}
	}

	return nil
}

// addCmdToTaskLog will manage the log of run commands in a file. This file gets append the path to a file
// for each command, that contains the executed script. The filename contains either ".done" or ".failed" as
// suffix, depending on the err given (nil or not).
func (build *Build) addCmdToTaskLog(tsk *task, checksumDir, checksum string, err error) (e error) {
	prefix := checksumDir + "/" + checksum
	sourceFile := prefix + ".sh"
	targetFile := prefix + ".done"
	if err != nil {
		logError(err)
		targetFile = prefix + ".failed"
	}
	rawCmd := fmt.Sprintf("{ [ -f %[1]s ] || mv %[2]s %[1]s; } && echo %[1]s >> %[3]s/%[4]s.run",
		targetFile, sourceFile, checksumDir, tsk.started.Format("20060102_150405"))
	c, e := build.prepareInternalCommand(rawCmd)
	if e != nil {
		return e
	}

	return c.Run()
}

type checksumTree map[string][]string

func (build *Build) buildChecksumTree() (ct checksumTree, e error) {
	ct = checksumTree{}

	rawCmd := fmt.Sprintf(
		`[ -d %[1]s ] && { ls %[1]s | while read dir; do ls -t %[1]s/$dir/*.run | head -n1 | xargs cat; done; }`,
		ukCACHEDIR)
	cmd, e := build.prepareInternalCommand(rawCmd)
	if e != nil {
		return nil, e
	}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}

	cmd.SetStdout(out)
	cmd.SetStderr(err)

	if e := cmd.Run(); e != nil {
		return nil, fmt.Errorf("%s: out=%s err=%s", e.Error(), out.String(), err.String())
	}

	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)

		if line == "" || !strings.HasSuffix(line, ".done") {
			continue
		}

		pkgname := filepath.Dir(strings.TrimPrefix(line, ukCACHEDIR+"/"))
		checksum := strings.TrimSuffix(filepath.Base(line), ".done")
		if len(checksum) != 64 {
			return nil, fmt.Errorf("invalid checksum %q found for package %q", checksum, pkgname)
		}
		ct[pkgname] = append(ct[pkgname], checksum)
	}

	return ct, nil
}

func (build *Build) prepareCommand(rawCmd string) (target.ExecCommand, error) {
	var sudo string
	if build.User() != "root" {
		sudo = "sudo "
	}
	return build.Command(sudo + rawCmd)
}

func (build *Build) prepareInternalCommand(rawCmd string) (target.ExecCommand, error) {
	rawCmd = fmt.Sprintf("sh -x -e <<\"EOC\"\n%s\nEOC\n", rawCmd)
	return build.prepareCommand(rawCmd)
}

func (build *Build) hostname() string {
	if s, ok := build.Target.(fmt.Stringer); ok {
		return s.String()
	}
	return "MISSING"
}

func (b *Build) commandAction(name string, checksums []string, c *commandWrapper) func() error {
	return func() error {
		s := struct {
			Command, Checksum, Name string
			ChecksumFiles           string
		}{
			Command:       c.command.Shell(),
			Checksum:      c.Checksum(),
			Name:          name,
			ChecksumFiles: strings.Join(checksums, "\n"),
		}
		cm, err := render(cmdTpl, s)
		if err != nil {
			return err
		}
		wg := &sync.WaitGroup{}
		defer wg.Wait()
		ec, err := b.Target.Command(cm)
		if err != nil {
			return err
		}
		o, err := ec.StdoutPipe()
		if err != nil {
			return err
		}
		e, err := ec.StderrPipe()
		if err != nil {
			return err
		}
		wg.Add(2)
		prefix := fmt.Sprintf("%s [%*s]", b.Target.String(), b.maxLength, name)
		go consumeStream(prefix, gocli.Red, e, wg)
		go consumeStream(prefix, func(in string) string { return in }, o, wg)
		fmt.Println(prefix + " " + c.LogMsg())
		if err := ec.Start(); err != nil {
			return err
		}
		return ec.Wait()
	}
}

func consumeStream(prefix string, form func(string) string, in io.Reader, wg *sync.WaitGroup) error {
	defer wg.Done()
	scanner := bufio.NewScanner(in)

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) > 2 {
			fmt.Printf("%s %s\n", prefix, form(strings.Join(fields[2:], "\t")))
		} else {
			fmt.Printf("%s %s\n", prefix, form(scanner.Text()))
		}
	}
	return scanner.Err()

}

func render(t string, i interface{}) (string, error) {
	tpl, err := template.New(t).Parse(t)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, i)
	return buf.String(), err
}

const cmdTpl = `set -e

function iso8601 {
    TZ=UTC date --iso-8601=ns | cut -d "+" -f 1
}

sudo_prefix=""
if [[ $(id -u) != 0 ]]; then
  sudo_prefix="sudo"
fi

build_date=$(TZ=utc date +"%Y%m%d_%H%M%S")
dir=/var/lib/urknall/{{ .Name }}/build.$build_date

$sudo_prefix mkdir -p $dir

$sudo_prefix tee $dir/{{ .Checksum }}.sh > /dev/null <<"UKEOF"
{{ .Command }}
UKEOF

done_path=$dir/{{ .Checksum }}.done
run_path=$dir/$build_date.run
log_path=$dir/{{ .Checksum }}.log
uk_path=/var/lib/urknall/{{ .Name }}

$sudo_prefix bash $dir/{{ .Checksum }}.sh 2> >(while read line; do echo "$(iso8601)	stderr	$line"; done | $sudo_prefix tee -a $log_path) > >(while read line; do echo "$(iso8601)	stdout	$line"; done | $sudo_prefix tee -a $log_path)

$sudo_prefix mv $dir/{{ .Checksum }}.sh $dir/{{ .Checksum }}.done
$sudo_prefix tee $run_path > /dev/null <<EOF
{{ .ChecksumFiles }}
EOF

$sudo_prefix mkdir -p $uk_path
$sudo_prefix cp $done_path $run_path $log_path $uk_path/
`

type taskState struct {
	name    string
	runSHAs []string
	content map[string]string
}

func readState(target Target) (content map[string]*taskState, err error) {
	cmd := `files=$(find /var/lib/urknall -maxdepth 1 -mindepth 1 -type d)

		if [[ -z $files ]]; then
		  exit
		fi

		tar cvz $(
			for dir in $files; do
				last_run=$(ls -t $dir/*.run | head -n1)
				echo $last_run
				cat $last_run
			done
		)
	`
	b, err := capture(target, cmd)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return map[string]*taskState{}, nil
	}

	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	t := tar.NewReader(gz)

	return readItemsFromTar(t)
}

func inspect(in interface{}) {
	b, err := json.MarshalIndent(in, "", "  ")
	if err == nil {
		os.Stdout.Write(b)
	}
}

func readItemsFromTar(t *tar.Reader) (m map[string]*taskState, err error) {
	m = map[string]*taskState{}
	for {
		switch h, err := t.Next(); err {
		case io.EOF:
			return m, nil
		case nil:
			name := filepath.Base(filepath.Dir(h.Name))
			b, err := ioutil.ReadAll(t)
			if err != nil {
				return nil, err
			}
			if _, ok := m[name]; !ok {
				m[name] = &taskState{content: map[string]string{}}
			}
			switch n := h.Name; {
			case strings.HasSuffix(n, ".run"):
				for _, f := range strings.Split(strings.TrimSpace(string(b)), "\n") {
					m[name].runSHAs = append(m[name].runSHAs, doneFileToChecksum(f))
				}
			case strings.HasSuffix(n, ".done"):
				m[name].content[doneFileToChecksum(n)] = strings.TrimSuffix(strings.TrimPrefix(string(b), "#!/bin/sh\nset -e\nset -x\n\n\n"), "\n")
			case strings.HasSuffix(n, ".log") || strings.HasSuffix(n, ".failed"):
				// ignore for now
			default:
				return nil, fmt.Errorf("%s dir=%t has unsupported suffix", n, h.FileInfo().IsDir())
			}
		default:
			return nil, err
		}
	}
}

func doneFileToChecksum(in string) string {
	return strings.TrimSuffix(filepath.Base(in), ".done")
}

func loadExecutedTasks(target Target, name string) (m map[string][]string, err error) {
	c, err := target.Command("ls -t /var/lib/urknall/" + name + "/*.run | head -n 1 | xargs cat")
	if err != nil {
		return nil, err
	}
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c.SetStderr(stdErr)
	c.SetStdout(stdOut)
	if err := c.Run(); err != nil {
		return nil, err
	}
	return nil, nil
}

func capture(target Target, cmd string) ([]byte, error) {
	c, err := target.Command(cmd)
	if err != nil {
		return nil, err
	}
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	c.SetStderr(stdErr)
	c.SetStdout(stdOut)
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("error running cmd=%q err=%q stderr=%q", cmd, err.Error(), stdErr.String())
	}
	return stdOut.Bytes(), nil
}

const (
	stateEcho = iota + 1
	stateDecoded
	stateMove
	statePostMove
)

func extractWriteFile(in string) (path, content string, ok bool, err error) {
	if !strings.Contains(in, "base64 -d | gunzip") {
		return "", "", false, nil
	}
	state := 0
	for _, f := range strings.Fields(in) {
		switch {
		case f == "echo":
			state++
		case state == stateEcho:
			content, err = unzip(f)
			if err != nil {
				return "", "", false, err
			}
			state++
		case f == "mv":
			state++
		case state == stateMove:
			state++
		case state == statePostMove:
			if path != "" {
				path += " "
			}
			path += f
		}
	}
	return path, content, true, nil
}

func unzip(f string) (content string, err error) {
	b, err := base64.StdEncoding.DecodeString(f)
	if err != nil {
		return "", err
	}
	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer gz.Close()
	b, err = ioutil.ReadAll(gz)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
