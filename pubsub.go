package urknall

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
	statusCached       = "CACHED"
	statusExecStart    = "EXEC"
	statusExecFinished = "FINISHED"
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
	key string

	dryRun     bool
	execStatus string
	cached     bool
	message    string
	host       *Host
	task       *taskData
	runlist    *Runlist

	publishedAt time.Time
	startedAt   time.Time

	duration               time.Duration
	totalRuntime           time.Duration
	sshResult              *gossh.Result
	line                   string
	stream                 string
	command                string
	invalidatedCachentries []string
	error_                 error
	stack                  string
}

// Predicated to verify whether the given message was sent via stderr.
func (message *Message) IsStderr() bool {
	return message.stream == "stderr"
}

// IP of the host the message was generated on. Will return an empty string if no host is specified (for example for
// urknall internal messages).
func (message *Message) HostIP() string {
	if message.host != nil {
		return message.host.IP
	}
	return ""
}

// Returns the name of the runlist that generated the message. Empty if no runlist specified.
func (message *Message) RunlistName() string {
	if message.runlist != nil {
		return message.runlist.name
	}
	return ""
}

// The key is an identifier for the message type. It's a point seperated path towards the message source.
func (message *Message) Key() string {
	return message.key
}

func (message Message) publishError(e error) {
	message.error_ = e
	message.publish("error")
}

func (message Message) publish(key string) {
	if message.key == "" {
		panic("message key must be set")
	}
	message.key += ("." + key)
	message.publishedAt = time.Now()
	if message.startedAt.IsZero() {
		message.startedAt = message.publishedAt
	} else {
		message.duration = message.publishedAt.Sub(message.startedAt)
	}

	publish(&message)
}
