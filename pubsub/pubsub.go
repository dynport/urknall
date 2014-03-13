package pubsub

import (
	"sync"
	"time"

	"github.com/dynport/dgtk/pubsub"
	"github.com/dynport/gossh"
)

var pubSub = []*pubsub.PubSub{}
var mutex = &sync.Mutex{}

// Register your own instance of the PubSub type to handle logging yourself.
func RegisterPubSub(ps *pubsub.PubSub) {
	mutex.Lock()
	pubSub = append(pubSub, ps)
	mutex.Unlock()
}

func publish(i interface{}) (e error) {
	for _, ps := range pubSub {
		if e = ps.Publish(i); e != nil {
			return e
		}
	}
	return nil
}

const (
	StatusCached       = "CACHED"
	StatusExecStart    = "EXEC"
	StatusExecFinished = "FINISHED"
)

const (
	MessageRunlistsPrecompile    = "urknall.runlists.precompile"
	MessageUrknallInternal       = "urknall.internal"
	MessageCleanupCacheEntries   = "urknall.cleanup_cache_entries"
	MessageRunlistsProvision     = "urknall.runlists.provision.list"
	MessageRunlistsProvisionTask = "urknall.runlists.provision.task"
)

// Urknall uses the github.com/dynport/dgtk/pubsub package for logging (a publisher-subscriber pattern where defined
// messages are sent to subscribers). This is the message type urknall will send out. If you handle logging yourself
// this type provides the required information.
type Message struct {
	Key string

	DryRun     bool
	ExecStatus string
	Cached     bool
	Message    string

	HostIP      string
	Task        string
	RunlistName string

	PublishedAt time.Time
	StartedAt   time.Time

	Duration               time.Duration
	TotalRuntime           time.Duration
	SshResult              *gossh.Result
	Line                   string
	Stream                 string
	Command                string
	InvalidatedCachentries []string
	Error_                 error
	Stack                  string
}

// Predicated to verify whether the given message was sent via stderr.
func (message *Message) IsStderr() bool {
	return message.Stream == "stderr"
}

func (message Message) PublishError(e error) {
	message.Error_ = e
	message.Publish("error")
}

func (message Message) Publish(key string) {
	if message.Key == "" {
		panic("message key must be set")
	}
	message.Key += ("." + key)
	message.PublishedAt = time.Now()
	if message.StartedAt.IsZero() {
		message.StartedAt = message.PublishedAt
	} else {
		message.Duration = message.PublishedAt.Sub(message.StartedAt)
	}

	publish(&message)
}
