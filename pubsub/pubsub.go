package pubsub

import (
	"runtime"
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

// Urknall uses the http://github.com/dynport/dgtk/pubsub package for logging (a publisher-subscriber pattern where
// defined messages are sent to subscribers). This is the message type urknall will send out. If you handle logging
// yourself this type provides the required information. Please note that this message is sent in different context's
// and not all fields will be set all the time.
type Message struct {
	Key string // Key the message is sent with.

	ExecStatus string // Urknall status (executed or cached).
	Message    string // The message to be logged.

	Hostname string // IP of the host a command is run.

	RunlistName  string // Name of the runlist currently being executed.
	TaskChecksum string // Hash of an runlist action.

	PublishedAt  time.Time     // When was the message published.
	StartedAt    time.Time     // When was the message created.
	Duration     time.Duration // How long did the action take (delta from message creation and publishing).
	TotalRuntime time.Duration // Timeframe of the action (might be larger than the message's).

	SshResult *gossh.Result // Result of an ssh call.

	Stream string // Stream a line appeared on.
	Line   string // Line that appeared on a stream.

	InvalidatedCacheEntries []string // List of invalidated cache entries (urknall caching).

	Error error  // Error that occured.
	Stack string // The stack trace in case of a panic.
}

// Predicated to verify whether the given message was sent via stderr.
func (message *Message) IsStderr() bool {
	return message.Stream == "stderr"
}

func (message Message) PublishError(e error) {
	message.Error = e
	message.Publish("error")
}

func (message *Message) PublishPanic(e error) {
	var buf []byte
	for read, size := 1024, 1024; read == size; read = runtime.Stack(buf, false) {
		buf = make([]byte, 2*size)
	}

	message.Stack = string(buf)
	message.Error = e
	message.Publish("panic")
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
