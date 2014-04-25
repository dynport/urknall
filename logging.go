package urknall

import (
	"time"

	"github.com/dynport/urknall/pubsub"
)

func message(key string, hostname string, rl *Package) (msg *pubsub.Message) {
	runlistName := ""
	if rl != nil {
		msg.RunlistName = runlistName
	}

	return &pubsub.Message{Key: key, StartedAt: time.Now(), Hostname: hostname, RunlistName: runlistName}
}
