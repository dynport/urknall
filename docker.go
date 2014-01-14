package urknall

import (
	"strings"
)

// Generate a Dockerfile from the commands collected on the runlist.
func (rl *Runlist) Dockerfile(from string) (string, error) {
	lines := []string{
		"FROM " + from,
	}
	for _, c := range rl.commands {
		lines = append(lines, c.Docker())
	}
	return strings.Join(lines, "\n"), nil
}
