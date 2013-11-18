package zwo

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/dynport/zwo/host"
	"os"
	"path/filepath"
	"strings"
)

type fileAction struct {
	path, content string
	owner         string
	mode          os.FileMode
	host          *host.Host
}

func (wCmd *fileAction) Docker() string {
	return "RUN " + wCmd.Shell()
}

var b64 = base64.StdEncoding

func (wCmd *fileAction) Shell() string {
	buf := &bytes.Buffer{}

	if wCmd.path == "" {
		panic("no path given")
	}

	if wCmd.content == "" {
		panic("no content given")
	}

	// Zip the content.
	zipper := gzip.NewWriter(buf)
	zipper.Write([]byte(wCmd.content))
	zipper.Flush()
	zipper.Close()

	// Encode the zipped content in Base64.
	encoded := b64.EncodeToString(buf.Bytes())

	// Compute sha256 hash of the encoded and zipped content.
	hash := sha256.New()
	hash.Write([]byte(wCmd.content))

	// Create temporary filename (hash as filename).
	tmpPath := fmt.Sprintf("/tmp/wunderscale.%x", hash.Sum(nil))

	// Get directory part of target file.
	dir := filepath.Dir(wCmd.path)

	// Create command, that will decode and unzip the content and write to the temporary file.
	cmd := ""
	cmd += fmt.Sprintf("mkdir -p %s", dir)
	cmd += fmt.Sprintf(" && echo %s | base64 -d | gunzip > %s", encoded, tmpPath)
	if wCmd.owner != "" { // If owner given, change accordingly.
		cmd += fmt.Sprintf(" && chown %s %s", wCmd.owner, tmpPath)
	}
	if wCmd.mode > 0 { // If mode given, change accordingly.
		cmd += fmt.Sprintf(" && chmod %o %s", wCmd.mode, tmpPath)
	}
	// Move the temporary file to the requested location.
	cmd += fmt.Sprintf(" && mv %s %s", tmpPath, wCmd.path)
	if wCmd.host.IsSudoRequired() {
		return fmt.Sprintf("sudo bash <<EOF\n%s\nEOF\n", cmd)
	}
	return cmd
}

func (wCmd *fileAction) Logging() string {
	sList := []string{"[FILE   ]"}

	if wCmd.host.IsSudoRequired() {
		sList = append(sList, "[SUDO]")
	}

	if wCmd.owner != "" && wCmd.owner != "root" {
		sList = append(sList, fmt.Sprintf("[CHOWN:%s]", wCmd.owner))
	}

	if wCmd.mode != 0 {
		sList = append(sList, fmt.Sprintf("[CHMOD:%.4o]", wCmd.mode))
	}

	sList = append(sList, wCmd.path)

	cLen := len(wCmd.content)
	if cLen > 50 {
		cLen = 50
	}
	sList = append(sList, fmt.Sprintf("<< %s", strings.Replace(wCmd.content[0:cLen], "\n", "⁋", -1)))
	return strings.Join(sList, " ")
}
