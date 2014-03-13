package urknall

import (
	"time"

	"github.com/dynport/urknall/pubsub"
)

func message(key string, host *Host, rl *Runlist) (msg *pubsub.Message) {
	msg = &pubsub.Message{Key: key, StartedAt: time.Now()}
	if host != nil {
		msg.HostIP = host.IP
	}
	if rl != nil {
		msg.RunlistName = rl.name
	}
	return msg
}
