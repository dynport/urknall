package urknall

import "github.com/dynport/urknall/cmd"

type commandWrapper struct {
	command cmd.Command
	cached  bool

	checksum string
	logMsg   string
}

func (cw *commandWrapper) Checksum() string {
	if cw.checksum == "" {
		var e error
		if cw.checksum, e = commandChecksum(cw.command); e != nil {
			panic(e)
		}
	}

	return cw.checksum
}

func (cw *commandWrapper) LogMsg() string {
	if logger, ok := cw.command.(cmd.Logger); ok {
		cw.logMsg = logger.Logging()
	} else {
		cw.logMsg = cw.command.Shell()
	}

	return cw.logMsg
}
