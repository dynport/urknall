package urknall

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dynport/urknall/cmd"
)

type remoteTaskRunner struct {
	build    *Build
	cmd      string
	dir      string
	command  cmd.Command
	taskName string

	started time.Time
}

func (runner *remoteTaskRunner) run() error {
	runner.started = time.Now()

	checksum, e := commandChecksum(runner.command)
	if e != nil {
		return e
	}
	prefix := runner.dir + "/" + checksum

	errors := make(chan error)
	logs := runner.newLogWriter(prefix+".log", errors)

	c, e := runner.build.prepareCommand(runner.cmd)
	if e != nil {
		return e
	}

	var wg sync.WaitGroup

	// Get pipes for stdout and stderr and forward messages to logs channel.
	stdout, e := c.StdoutPipe()
	if e != nil {
		return e
	}
	wg.Add(1)
	go runner.forwardStream(logs, "stdout", &wg, stdout)

	stderr, e := c.StderrPipe()
	if e != nil {
		return e
	}
	wg.Add(1)
	go runner.forwardStream(logs, "stderr", &wg, stderr)

	e = c.Run()
	wg.Wait()
	close(logs)

	runner.writeChecksumFile(prefix, e)
	// TODO(gf): should add better error handling (don't panic right now, but properly close everything open).

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
	cmd := fmt.Sprintf("echo %q > %s && echo %s >> %s/%s.run", runner.command.Shell(), targetFile, targetFile, runner.dir, runner.started.Format("20060102_150405"))
	c, e := runner.build.prepareCommand(cmd)
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

func (runner *remoteTaskRunner) forwardStream(logs chan string, stream string, wg *sync.WaitGroup, r io.Reader) {
	defer wg.Done()

	m := message("task.io", runner.build.hostname(), runner.taskName)
	m.Message = runner.command.Shell()
	if logger, ok := runner.command.(cmd.Logger); ok {
		m.Message = logger.Logging()
	}
	m.Stream = stream

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		m.Line = scanner.Text()
		m.TotalRuntime = time.Since(runner.started)
		m.Publish(stream)
		logs <- time.Now().UTC().Format(time.RFC3339Nano) + "\t" + stream + "\t" + scanner.Text()
	}
}

func (runner *remoteTaskRunner) newLogWriter(path string, errors chan error) chan string {
	logs := make(chan string)
	go func() {
		// so ugly, but: sudo not required and "sh -c" adds some escaping issues with the variables. This is why Command is called directly.
		c, e := runner.build.Command("cat - > " + path)
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
			if _, e = io.WriteString(in, log+"\n"); e != nil {
				errors <- e
			}
		}

		if in, ok := in.(io.WriteCloser); ok {
			if e = in.Close(); e != nil {
				errors <- e
			}
		}

		// Close the stdin pipe of the above command (terminating that).
		// Wait for above command to return.
		errors <- c.Wait()
	}()
	return logs
}
