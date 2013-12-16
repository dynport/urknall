package urknall

import (
	"github.com/dynport/dgtk/pubsub"
	"time"
)

var pubSub = &pubsub.PubSub{}

func Publish(key string, i interface{}) {
	pubSub.Publish(key, i)
}

func Subscribe(pattern string, c chan *pubsub.Message) {
	pubSub.Subscribe(pattern, c)
}

const (
	StatusCached       = "CACHED  "
	StatusExecStart    = "EXEC    "
	StatusExecFinished = "FINISHED"
)

type Message struct {
	DryRun                 bool
	ExecStatus             string
	Cached                 bool
	Message                string
	Host                   *Host
	Task                   *taskData
	Runlist                *Runlist
	Duration               time.Duration
	IOMessage              []interface{}
	Stream                 string
	Command                string
	InvalidatedCachentries []string
	Error                  error
	Stack                  string
}
