// Package urknall
//
// See http://urknall.dynport.de for detailed documentation.
package urknall

import (
	"io"

	"github.com/dynport/urknall/pubsub"
)

// OpenLogger creates a logging facility for urknall using the given writer for
// output. Note that the resource must be closed!
func OpenLogger(w io.Writer) io.Closer {
	return pubsub.OpenLogger(w)
}
