package urknall

import (
	"bufio"
	"io"
	"log"
	"strings"
	"time"
)

type remoteTaskRunner struct {
	Runner   *Runner
	cmd      string
	dir      string
	task     *taskData
	hostname string

	started time.Time
}

func (runner *remoteTaskRunner) baseCommand() string {
	cmd := strings.Join(runner.Runner.Env, " ") + " bash -s -x -e"
	if runner.Runner.IsSudoRequired() {
		cmd = "sudo " + cmd
	}
	return cmd
}

func (runner *remoteTaskRunner) run() error {
	runner.started = time.Now()

	prefix := runner.dir + "/" + runner.task.checksum
	errors := make(chan error)
	logs := runner.newLogWriter(prefix+".log", errors)

	c, e := runner.Runner.Commander.Command(runner.cmd)
	if e != nil {
		return e
	}

	// Get pipes for stdout and stderr and forward messages to logs channel.
	stdout, e := c.StdoutPipe()
	if e != nil {
		return e
	}
	finishedMap := map[string]interface{}{
		"stdout": true,
		"stderr": true,
	}
	finishedChannel := make(chan string)
	go runner.forwardStream(logs, "stdout", stdout, finishedChannel)

	stderr, e := c.StderrPipe()
	if e != nil {
		return e
	}
	go runner.forwardStream(logs, "stderr", stderr, finishedChannel)

	e = c.Run()
	// Command was executed. Close the logging channel (thereby closing the back-channel of the logs).
	for len(finishedMap) > 0 {
		select {
		case s := <-finishedChannel:
			delete(finishedMap, s)
		}
	}
	close(logs)

	runner.writeChecksumFile(prefix, e)

	// Get errors that might have occured while handling the back-channel for the logs.
	select {
	case e := <-errors:
		if e != nil {
			log.Printf("ERROR: %s", e.Error())
		}
	}
	return e
}

func (runner *remoteTaskRunner) writeChecksumFile(prefix string, e error) {
	targetFile := prefix + ".done"
	if e != nil {
		logError(e)
		targetFile = prefix + ".failed"
	}
	cmd := "cat > " + targetFile + " <<EOF\n" + runner.task.Command().Shell() + "\nEOF"
	c, e := runner.Runner.Commander.Command(cmd)
	if e != nil {
		panic(e.Error())
	}

	if e := c.Run(); e != nil {
		panic(e.Error())
	}
}

func logError(e error) {
	log.Printf("ERROR: %s", e.Error())
}

func (runner *remoteTaskRunner) forwardStream(logs chan string, stream string, r io.Reader, finished chan string) {
	m := message("task.io", runner.Runner.Hostname(), runner.task.runlist)
	m.Message = runner.task.Command().Logging()
	m.Stream = stream

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		m.Line = line
		m.TotalRuntime = time.Since(runner.started)
		m.Publish(stream)
		logs <- time.Now().UTC().Format(time.RFC3339Nano) + "\t" + stream + "\t" + scanner.Text()
	}
	finished <- stream
}

func (runner *remoteTaskRunner) newLogWriter(path string, errors chan error) chan string {
	logs := make(chan string)
	go func() {
		c, e := runner.Runner.Commander.Command("{ t=$(tempfile -m0660) || exit 1; } && cat - > $t && mv $t " + path + " && chgrp urknall " + path)
		if e != nil {
			errors <- e
			return
		}

		// Get pipe to stdin of the execute command.
		in, e := c.StdinPipe()
		if e != nil {
			errors <- e
			return
		}

		// Run command, writing everything coming from stdin to a file.

		e = c.Start()
		if e != nil {
			errors <- e
			return
		}

		// Send all messages from logs to the stdin of the new session.
		for log := range logs {
			io.WriteString(in, log+"\n")
		}

		if in, ok := in.(io.WriteCloser); ok {
			in.Close()
		}

		// Close the stdin pipe of the above command (terminating that).
		// Wait for above command to return.
		errors <- c.Wait()
	}()
	return logs
}
