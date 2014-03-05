package urknall

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
)

type remoteTaskRunner struct {
	dir        string
	task       *taskData
	host       *Host
	clientConn *ssh.ClientConn

	started time.Time
}

func (runner *remoteTaskRunner) baseCommand() string {
	cmd := strings.Join(runner.host.Env, " ") + " bash -s -x -e"
	if runner.host.isSudoRequired() {
		cmd = "sudo " + cmd
	}
	return cmd
}

func (runner *remoteTaskRunner) run() error {
	runner.started = time.Now()

	prefix := runner.dir + "/" + runner.task.checksum
	errors := make(chan error)
	logs := runner.newLogWriter(prefix+".log", errors)

	session, e := runner.clientConn.NewSession()
	if e != nil {
		return e
	}
	defer session.Close()

	// Get pipes for stdout and stderr and forward messages to logs channel.
	stdout, e := session.StdoutPipe()
	if e != nil {
		return e
	}
	finishedMap := map[string]interface{}{
		"stdout": true,
		"stderr": true,
	}
	finishedChannel := make(chan string)
	go runner.forwardStream(logs, "stdout", stdout, finishedChannel)

	stderr, e := session.StderrPipe()
	if e != nil {
		return e
	}
	go runner.forwardStream(logs, "stderr", stderr, finishedChannel)

	stdin, e := session.StdinPipe()
	if e != nil {
		return e
	}

	e = session.Start(runner.baseCommand())

	cmd := runner.task.Command().Shell()
	if n, e := io.WriteString(stdin, cmd); e != nil {
		return e
	} else if n != len(cmd) {
		return fmt.Errorf("failed to write complete string")
	}
	stdin.Close()

	if e = session.Wait(); e != nil {
		return e
	}

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
	_ = executeCommand(runner.clientConn, cmd)
}

func executeCommand(con *ssh.ClientConn, cmd string) error {
	ses, e := con.NewSession()
	if e != nil {
		return e
	}
	defer ses.Close()
	buf := &bytes.Buffer{}
	ses.Stderr = buf
	e = ses.Run(cmd)
	if e != nil {
		return fmt.Errorf(e.Error() + ": " + buf.String())
	}
	return nil
}

func logError(e error) {
	log.Printf("ERROR: %s", e.Error())
}

func (runner *remoteTaskRunner) forwardStream(logs chan string, stream string, r io.Reader, finished chan string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		m := &Message{key: "task.io", host: runner.host, stream: stream, task: runner.task, line: line, runlist: runner.task.runlist, totalRuntime: time.Since(runner.started)}
		m.publish(stream)
		logs <- time.Now().UTC().Format(time.RFC3339Nano) + "\t" + stream + "\t" + scanner.Text()
	}
	finished <- stream
}

func (runner *remoteTaskRunner) newLogWriter(path string, errors chan error) chan string {
	logs := make(chan string)
	go func() {
		ses, e := runner.clientConn.NewSession()
		if e != nil {
			errors <- e
			return
		}
		defer ses.Close()

		// Get pipe to stdin of the execute command.
		in, e := ses.StdinPipe()
		if e != nil {
			errors <- e
			return
		}

		// Run command, writing everything coming from stdin to a file.
		ses.Start("{ t=$(tempfile -m0660) || exit 1; } && cat - > $t && mv $t " + path + " && chgrp urknall " + path)

		// Send all messages from logs to the stdin of the new session.
		for log := range logs {
			io.WriteString(in, log+"\n")
		}

		// Close the stdin pipe of the above command (terminating that).
		in.Close()
		// Wait for above command to return.
		errors <- ses.Wait()
	}()
	return logs
}
