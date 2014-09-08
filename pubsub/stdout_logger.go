package pubsub

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dynport/dgtk/pubsub"
)

const (
	colorDryRun = 226
	colorCached = 33
	colorExec   = 34
)

var colorMapping = map[string]int{
	StatusCached:       colorCached,
	StatusExecFinished: colorExec,
}

var ignoredMessagesError = errors.New("ignored published messages (subscriber buffer full)")

// Create a logging facility for urknall using urknall's default formatter.
// Note that this resource must be closed afterwards!
func OpenLogger(w io.Writer) io.Closer {
	logger := &logger{}
	logger.Output = w
	logger.Formatter = logger.DefaultFormatter
	// Ignore the error from Start. It would only be triggered if the formatter wouldn't be set.
	_ = logger.Start()
	return logger
}

type logger struct {
	Output       io.Writer
	Formatter    formatter
	maxLengths   map[int]int
	started      time.Time
	finished     chan interface{}
	pubSub       *pubsub.PubSub
	subscription *pubsub.Subscription
}

func (logger *logger) Started() time.Time {
	if logger.started.IsZero() {
		logger.started = time.Now()
	}
	return logger.started
}

func (logger *logger) formatCommandOuput(message *Message) string {
	prefix := fmt.Sprintf("[%s][%s][%s]", formatIp(message.Hostname), formatRunlistName(message.RunlistName, 12), formatDuration(logger.sinceStarted()))
	line := message.Line
	if message.IsStderr() {
		line = colorize(1, line)
	}
	return prefix + " " + line
}

func formatIp(ip string) string {
	return fmt.Sprintf("%15s", ip)
}

type formatter func(urknallMessage *Message) string

func (logger *logger) DefaultFormatter(message *Message) string {
	ignoreKeys := []string{MessageRunlistsPrecompile, MessageCleanupCacheEntries, MessageRunlistsProvision, MessageUrknallInternal}
	for _, k := range ignoreKeys {
		if strings.HasPrefix(message.Key, k) {
			return ""
		}
	}
	if len(message.Line) > 0 {
		return logger.formatCommandOuput(message)
	}
	ip := message.Hostname
	runlistName := message.RunlistName
	payload := ""
	if message.Message != "" {
		payload = message.Message
	}
	execStatus := fmt.Sprintf("%-8s", message.ExecStatus)
	if color := colorMapping[message.ExecStatus]; color > 0 {
		execStatus = colorize(color, execStatus)
	}
	parts := []string{
		fmt.Sprintf("[%s][%s][%s][%s]%s",
			formatIp(ip),
			formatRunlistName(runlistName, 12),
			formatDuration(logger.sinceStarted()),
			execStatus,
			payload,
		),
	}
	return strings.Join(parts, " ")
}

func formatRunlistName(name string, maxLen int) string {
	if len(name) > maxLen {
		name = name[0:maxLen]
	}
	return fmt.Sprintf("%-*s", maxLen, name)
}

func formatDuration(dur time.Duration) string {
	durString := ""
	if dur >= 1*time.Millisecond {
		durString = fmt.Sprintf("%.03f", dur.Seconds())
	}
	return fmt.Sprintf("%7s", durString)
}

func (logger *logger) sinceStarted() time.Duration {
	return time.Now().Sub(logger.Started())
}

func (logger *logger) Start() error {
	logger.started = time.Now()
	if logger.Formatter == nil {
		return fmt.Errorf("Formatter must be set")
	}
	logger.pubSub = pubsub.New()
	RegisterPubSub(logger.pubSub)
	logger.subscription = logger.pubSub.Subscribe(func(m *Message) {
		if message := logger.Formatter(m); message != "" {
			fmt.Fprintln(logger.Output, message)
		}
	})
	return nil
}

func (logger *logger) Close() (e error) {
	e = logger.subscription.Close()
	if d := logger.pubSub.Stats.Ignored(); e == nil && d > 0 {
		return ignoredMessagesError
	}
	return e
}

func colorize(c int, s string) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", c, s)
}
