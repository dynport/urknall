package urknall

import (
	"github.com/dynport/dgtk/pubsub"
	"github.com/dynport/gossh"
	"sync"
	"time"
)

var pubSub = []*pubsub.PubSub{}
var mutex = &sync.Mutex{}

func RegisterPubSub(ps *pubsub.PubSub) {
	mutex.Lock()
	pubSub = append(pubSub, ps)
	mutex.Unlock()
}

func publish(i interface{}) {
	for _, ps := range pubSub {
		ps.Publish(i)
	}
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

func (message *Message) IsStderr() bool {
	return message.stream == "stderr"
}

func (message *Message) HostIP() string {
	if message.host != nil {
		return message.host.IP
	}
	return ""
}

func (message *Message) RunlistName() string {
	if message.runlist != nil {
		return message.runlist.name
	}
	return ""
}

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
