package logging

import (
	"fmt"
	"github.com/dynport/dgtk/pubsub"
	"github.com/dynport/gocli"
	"github.com/dynport/urknall"
	"log"
	"strings"
	"time"
)

const (
	ColorDryRun = 226
	ColorCached = 33
	ColorExec   = 34
)

var colorMapping = map[string]int{
	urknall.StatusCached:       ColorCached,
	urknall.StatusExecFinished: ColorExec,
}

type StdoutLogger struct {
	Formatter  Formatter
	maxLengths map[int]int
	started    time.Time
	channel    chan *pubsub.Message
	finished   chan interface{}
}

func (logger *StdoutLogger) Started() time.Time {
	if logger.started.IsZero() {
		logger.started = time.Now()
	}
	return logger.started
}

func (logger *StdoutLogger) formatCommandOuput(message *urknall.Message) string {
	runlist := ""
	if message.Runlist != nil {
		runlist = message.Runlist.Name()
	}
	prefix := fmt.Sprintf("[%s][%-8s][%s]", formatIp(message.Host.IP), runlist, formatDuration(logger.SinceStarted()))
	text := fmt.Sprint(message.IOMessage...)
	if message.Stream == "stderr" {
		text = gocli.Red(text)
	}
	return prefix + " " + text
}

func formatIp(ip string) string {
	return fmt.Sprintf("%15s", ip)
}

type Formatter func(m *pubsub.Message, urknallMessage *urknall.Message) string

func (logger *StdoutLogger) DefaultFormatter(m *pubsub.Message, urknallMessage *urknall.Message) string {
	if strings.HasPrefix(m.Key(), "runlists.precompiling") {
		return ""
	}
	if strings.HasPrefix(m.Key(), "urknall.cleanup_cache_entries") {
		return ""
	}
	if len(urknallMessage.IOMessage) > 0 {
		return logger.formatCommandOuput(urknallMessage)
	}
	ip := ""
	if urknallMessage.Host != nil {
		ip = urknallMessage.Host.IP
	}
	runlistName := ""
	if urknallMessage.Runlist != nil {
		runlistName = urknallMessage.Runlist.Name()
	}
	payload := ""
	if urknallMessage.Task != nil {
		payload = urknallMessage.Task.Command().Logging()
	}
	execStatus := urknallMessage.ExecStatus
	if color := colorMapping[execStatus]; color > 0 {
		execStatus = gocli.Colorize(color, execStatus)
	}
	parts := []string{
		fmt.Sprintf("[%s][%-8s][%s][%s][%-8s] %s",
			formatIp(ip),
			runlistName,
			formatDuration(urknallMessage.Duration),
			formatDuration(logger.SinceStarted()),
			execStatus,
			payload,
		),
	}
	return strings.Join(parts, " ")
}

func formatDuration(dur time.Duration) string {
	durString := ""
	if dur >= 1*time.Millisecond {
		durString = fmt.Sprintf("%.03f", dur.Seconds())
	}
	return fmt.Sprintf("%7s", durString)
}

func (logger *StdoutLogger) SinceStarted() time.Duration {
	return time.Now().Sub(logger.Started())
}

func (logger *StdoutLogger) Start() error {
	logger.started = time.Now()
	if logger.Formatter == nil {
		return fmt.Errorf("Formatter must be set")
	}
	logger.channel = make(chan *pubsub.Message, 1000)
	logger.finished = make(chan interface{})
	go func(c chan *pubsub.Message, finished chan interface{}) {
		for m := range c {
			if message, ok := m.Payload().(*urknall.Message); ok {
				if message := logger.Formatter(m, message); message != "" {
					fmt.Println(message)
				}
			}
		}
		finished <- nil
	}(logger.channel, logger.finished)
	urknall.Subscribe("*", logger.channel)
	return nil
}

func (logger *StdoutLogger) Close() {
	close(logger.channel)
	timer := time.NewTimer(5 * time.Second)
	select {
	case <-logger.finished:
	case <-timer.C:
		log.Print("ERROR: logger did not finish up")
	}
}

func init() {
	log.SetFlags(0)
}
