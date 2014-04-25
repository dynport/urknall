package urknall

import (
	"time"

	"github.com/dynport/urknall/pubsub"
)

func message(key string, hostname string, rl *Package) (msg *pubsub.Message) {
	msg = &pubsub.Message{Key: key, StartedAt: time.Now()}
	if hostname != "" {
		msg.HostIP = hostname
	}
	if rl != nil {
		msg.RunlistName = rl.name
	}
	return msg
}
