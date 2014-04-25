package urknall

import (
	"time"

	"github.com/dynport/urknall/pubsub"
)

func message(key string, hostname string, pkg *Package) (msg *pubsub.Message) {
	runlistName := ""
	if pkg != nil {
		runlistName = pkg.name
	}

	return &pubsub.Message{Key: key, StartedAt: time.Now(), Hostname: hostname, RunlistName: runlistName}
}
