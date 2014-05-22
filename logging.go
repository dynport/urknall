package urknall

import (
	"time"

	"github.com/dynport/urknall/pubsub"
)

func message(key string, hostname string, taskName string) (msg *pubsub.Message) {
	runlistName := ""
	if taskName != "" {
		runlistName = taskName
	}

	return &pubsub.Message{Key: key, StartedAt: time.Now(), Hostname: hostname, RunlistName: runlistName}
}
