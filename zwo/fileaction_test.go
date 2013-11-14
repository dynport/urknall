package zwo

import (
	"fmt"
	"github.com/dynport/zwo/host"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteFile(t *testing.T) {
	h, e := host.NewHost(host.HOST_TYPE_SSH)
	assert.Nil(t, e)

	rawContent := "something"
	zippedContent := `H4sIAAAJbogA/yrOz00tycjMSwcAAAD//wEAAP//+zHaCQkAAAA=`
	hash := "3fc9b689459d738f8c88a3a48aa9e33542016b7a4052e001aaa536fca74813cb"
	tmpFile := fmt.Sprintf("/tmp/wunderscale.%s", hash)

	cmd, e := WriteFile("", "", "", 0)(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty path given")

	cmd, e = WriteFile("/foo", "", "", 0)(h, &struct{}{})
	assert.Nil(t, cmd)
	assert.Error(t, e, "empty content given")

	commandBase := fmt.Sprintf("mkdir -p / && echo %s | base64 -d | gunzip > %s", zippedContent, tmpFile)

	cmd, e = WriteFile("/foo", rawContent, "", 0)(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), fmt.Sprintf("%s && mv %s /foo", commandBase, tmpFile))
	assert.Nil(t, e)

	cmd, e = WriteFile("/foo", rawContent, "nobody", 0)(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), fmt.Sprintf("%s && chown %s %s && mv %s /foo", commandBase, "nobody", tmpFile, tmpFile))
	assert.Nil(t, e)

	cmd, e = WriteFile("/foo", "something", "nobody", 0666)(h, &struct{}{})
	assert.Equal(t, cmd.Plain(), fmt.Sprintf("%s && chown %s %s && chmod %0o %s && mv %s /foo", commandBase, "nobody", tmpFile, 0666, tmpFile, tmpFile))
	assert.Nil(t, e)
}
