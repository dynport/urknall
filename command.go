package urknall

import "github.com/dynport/urknall/utils"

type stringCommand struct {
	cmd string
}

func (sc *stringCommand) Shell() string {
	return sc.cmd
}

func (sc *stringCommand) Logging() string {
	return "[COMMAND] " + sc.cmd
}

func (sc *stringCommand) Render(i interface{}) {
	sc.cmd = utils.MustRenderTemplate(sc.cmd, i)
}
