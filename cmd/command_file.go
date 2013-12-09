package cmd

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/dynport/zwo/assets"
	"github.com/dynport/zwo/host"
	"os"
	"path/filepath"
	"strings"
)

// The "FileCommand" is used to write files to the host being provisioned. The go templating mechanism
// (see http://golang.org/pkg/text/template) is used with the file's content and the package the command for rendering
// the file is taken from. Thereby it is possible to have dynamic content (based on the packages' configuration) for the
// file content and at the same time store it in an asset (that must not be preprocessed or compiled by hand).
type FileCommand struct {
	Path        string      // Path to the file to create.
	Content     string      // Content of the file to create.
	Owner       string      // Owner of the file to create (root per default).
	Permissions os.FileMode // Permissions of the file created (only changed from system default if set).
}

// Helper method to create a file at the given path with the given content, and with owner and permissions set
// accordingly.
func WriteFile(path string, content string, owner string, permissions os.FileMode) *FileCommand {
	return &FileCommand{Path: path, Content: content, Owner: owner, Permissions: permissions}
}

// Helper method to write the asset with the given name to the location given, with owner and permissions set
// accordingly. If no asset with the given name exists the function will panic!
func WriteAsset(path, assetName, owner string, permissions os.FileMode) *FileCommand {
	content, e := assets.Get(assetName)
	if e != nil {
		panic(e)
	}
	return WriteFile(path, string(content), owner, permissions)
}

func (fc *FileCommand) Docker(host *host.Host) string {
	return "RUN " + fc.Shell(host)
}

var b64 = base64.StdEncoding

func (fc *FileCommand) Shell(host *host.Host) string {
	buf := &bytes.Buffer{}

	if fc.Path == "" {
		panic("no path given")
	}

	if fc.Content == "" {
		panic("no content given")
	}

	// Zip the content.
	zipper := gzip.NewWriter(buf)
	zipper.Write([]byte(fc.Content))
	zipper.Flush()
	zipper.Close()

	// Encode the zipped content in Base64.
	encoded := b64.EncodeToString(buf.Bytes())

	// Compute sha256 hash of the encoded and zipped content.
	hash := sha256.New()
	hash.Write([]byte(fc.Content))

	// Create temporary filename (hash as filename).
	tmpPath := fmt.Sprintf("/tmp/wunderscale.%x", hash.Sum(nil))

	// Get directory part of target file.
	dir := filepath.Dir(fc.Path)

	// Create command, that will decode and unzip the content and write to the temporary file.
	cmd := ""
	cmd += fmt.Sprintf("mkdir -p %s", dir)
	cmd += fmt.Sprintf(" && echo %s | base64 -d | gunzip > %s", encoded, tmpPath)
	if fc.Owner != "" { // If owner given, change accordingly.
		cmd += fmt.Sprintf(" && chown %s %s", fc.Owner, tmpPath)
	}
	if fc.Permissions > 0 { // If mode given, change accordingly.
		cmd += fmt.Sprintf(" && chmod %o %s", fc.Permissions, tmpPath)
	}
	// Move the temporary file to the requested location.
	cmd += fmt.Sprintf(" && mv %s %s", tmpPath, fc.Path)
	return cmd
}

func (fc *FileCommand) Logging(host *host.Host) string {
	sList := []string{"[FILE   ]"}

	if fc.Owner != "" && fc.Owner != "root" {
		sList = append(sList, fmt.Sprintf("[CHOWN:%s]", fc.Owner))
	}

	if fc.Permissions != 0 {
		sList = append(sList, fmt.Sprintf("[CHMOD:%.4o]", fc.Permissions))
	}

	sList = append(sList, " "+fc.Path)

	cLen := len(fc.Content)
	if cLen > 50 {
		cLen = 50
	}
	sList = append(sList, fmt.Sprintf(" << %s", strings.Replace(string(fc.Content[0:cLen]), "\n", "⁋", -1)))
	return strings.Join(sList, "")
}
