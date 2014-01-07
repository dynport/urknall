package urknall

import (
	"strings"
)

func (rl *Runlist) Dockerfile(from string) (string, error) {
	lines := []string{
		"FROM " + from,
	}
	for _, c := range rl.commands {
		lines = append(lines, c.Docker())
	}
	return strings.Join(lines, "\n"), nil
}
