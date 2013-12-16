package urknall

import (
	"fmt"
	"github.com/dynport/dgtk/pubsub"
	"github.com/dynport/gocli"
	"io"
	"log"
	"strings"
	"time"
)

const (
	colorDryRun = 226
	colorCached = 33
	colorExec   = 34
)

var colorMapping = map[string]int{
	statusCached:       colorCached,
	statusExecFinished: colorExec,
}

// needs to be closed afterwards
func OpenStdoutLogger() (io.Closer, error) {
	logger := &stdoutLogger{}
	logger.Formatter = logger.DefaultFormatter
	e := logger.Start()
	if e != nil {
		return nil, e
	}
	return logger, nil
}

type stdoutLogger struct {
	Formatter    formatter
	maxLengths   map[int]int
	started      time.Time
	finished     chan interface{}
	pubSub       *pubsub.PubSub
	subscription *pubsub.Subscription
}

func (logger *stdoutLogger) Started() time.Time {
	if logger.started.IsZero() {
		logger.started = time.Now()
	}
	return logger.started
}

func (logger *stdoutLogger) formatCommandOuput(message *Message) string {
	prefix := fmt.Sprintf("[%s][%-8s][%s]", formatIp(message.HostIP()), message.RunlistName(), formatDuration(logger.sinceStarted()))
	text := fmt.Sprint(message.iOMessages...)
	if message.IsStderr() {
		text = gocli.Red(text)
	}
	return prefix + " " + text
}

func formatIp(ip string) string {
	return fmt.Sprintf("%15s", ip)
}

type formatter func(urknallMessage *Message) string

func (logger *stdoutLogger) DefaultFormatter(message *Message) string {
	ignoreKeys := []string{MessageRunlistsPrecompile, MessageCleanupCacheEntries, MessageRunlistsProvision, MessageUrknallInternal}
	for _, k := range ignoreKeys {
		if strings.HasPrefix(message.Key(), k) {
			return ""
		}
	}
	if len(message.iOMessages) > 0 {
		return logger.formatCommandOuput(message)
	}
	ip := message.HostIP()
	runlistName := message.RunlistName()
	payload := ""
	if message.task != nil {
		payload = message.task.Command().Logging()
	}
	execStatus := message.execStatus
	if color := colorMapping[execStatus]; color > 0 {
		execStatus = gocli.Colorize(color, execStatus)
	}
	parts := []string{
		fmt.Sprintf("[%s][%-8s][%s][%-8s] %s",
			formatIp(ip),
			runlistName,
			formatDuration(logger.sinceStarted()),
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

func (logger *stdoutLogger) sinceStarted() time.Duration {
	return time.Now().Sub(logger.Started())
}

func (logger *stdoutLogger) Start() error {
	logger.started = time.Now()
	if logger.Formatter == nil {
		return fmt.Errorf("Formatter must be set")
	}
	logger.pubSub = pubsub.New()
	RegisterPubSub(logger.pubSub)
	logger.subscription = logger.pubSub.Subscribe(func(m *Message) {
		if message := logger.Formatter(m); message != "" {
			log.Println(message)
		}
	})
	return nil
}

func (logger *stdoutLogger) Close() error {
	return logger.subscription.Close()
}

func init() {
	log.SetFlags(0)
}
