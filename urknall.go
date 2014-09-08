// The urknall library: see http://urknall.dynport.de/docs/library/ for further information.
package urknall

import (
	"io"

	"github.com/dynport/urknall/pubsub"
)

// Create a logging facility for urknall using the given writer for output.
func OpenLogger(w io.Writer) io.Closer {
	return pubsub.OpenLogger(w)
}
