// The Urknall Library
//
// This is the urknall library. The core action it provides is building a
// target with a package. The target could be a host reached via SSH for
// example. A package contains tasks, containing the commands executed on
// the target. Each task has a caching layer, that prevents commands from being
// executed if they have been executed before and none of the preceding
// commands has changed.
//
// A quick example can be found at the Build type documentation. A more real
// life example is located in the example directory.
//
// TODO(gf): Add reference to the website with more general information.
package urknall

import (
	"io"

	"github.com/dynport/urknall/pubsub"
)

// Create a logging facility for urknall using the given writer for output.
// Note that this resource must be closed afterwards! The following pattern is
// idiomatic:
//   defer urknall.OpenLogger(os.Stdout).Close()
func OpenLogger(w io.Writer) io.Closer {
	return pubsub.OpenLogger(w)
}
