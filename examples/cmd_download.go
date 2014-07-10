package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dynport/urknall/utils"
)

const TMP_DOWNLOAD_DIR = "/tmp/downloads"

// Download the URL and write the file to the given destination, with owner and permissions set accordingly.
// Destination can either be an existing directory or a file. If a directory is given the downloaded file will moved
// there using the file name from the URL. If it is a file, the downloaded file will be moved (and possibly renamed) to
// that destination. If the extract flag is set the downloaded file will be extracted to the directory given in the
// destination field.
type DownloadCommand struct {
	Url         string      // Where to download from.
	Destination string      // Where to put the downloaded file.
	Owner       string      // Owner of the downloaded file.
	Permissions os.FileMode // Permissions of the downloaded file.
	Extract     bool        // Extract the downloaded archive.
}

// Download the file from the given URL and extract it to the given directory. If the directory does not exist it is
// created. See the "ExtractFile" command for a list of supported archive types.
func DownloadAndExtract(url, destination string) *DownloadCommand {
	return &DownloadCommand{Url: url, Destination: destination, Extract: true}
}

func Download(url, destination, owner string, permissions os.FileMode) *DownloadCommand {
	return &DownloadCommand{Url: url, Destination: destination, Owner: owner, Permissions: permissions}
}

func (cmd *DownloadCommand) Validate() error {
	if cmd.Url == "" {
		return fmt.Errorf("Url must be set")
	}
	if cmd.Destination == "" {
		return fmt.Errorf("Destination to download %q to must be set", cmd.Url)
	}
	return nil
}

func (cmd *DownloadCommand) Render(i interface{}) {
	cmd.Url = utils.MustRenderTemplate(cmd.Url, i)
	cmd.Destination = utils.MustRenderTemplate(cmd.Destination, i)
}

func (dc *DownloadCommand) Shell() string {
	filename := path.Base(dc.Url)
	destination := fmt.Sprintf("%s/%s", TMP_DOWNLOAD_DIR, filename)

	cmd := []string{}

	cmd = append(cmd, fmt.Sprintf("which curl > /dev/null || { apt-get update && apt-get install -y curl; }"))
	cmd = append(cmd, fmt.Sprintf("mkdir -p %s", TMP_DOWNLOAD_DIR))
	cmd = append(cmd, fmt.Sprintf("cd %s", TMP_DOWNLOAD_DIR))
	cmd = append(cmd, fmt.Sprintf(`curl -SsfLO "%s"`, dc.Url))

	switch {
	case dc.Extract && dc.Destination == "":
		panic(fmt.Errorf("shall extract, but don't know where (i.e. destination field is empty"))
	case dc.Extract:
		cmd = append(cmd, Extract(destination, dc.Destination).Shell())
	case dc.Destination != "":
		cmd = append(cmd, fmt.Sprintf("mv %s %s", destination, dc.Destination))
		destination = dc.Destination
	}

	if dc.Owner != "" && dc.Owner != "root" {
		ifFile := fmt.Sprintf("{ if [ -f %s ]; then chown %s %s; fi; }", destination, dc.Owner, destination)
		ifInDir := fmt.Sprintf("{ if [ -d %s && -f %s/%s ]; then chown %s %s/%s; fi; }", destination, destination, filename, dc.Owner, destination, filename)
		ifDir := fmt.Sprintf("{ if [ -d %s ]; then chown -R %s %s; fi; }", destination, dc.Owner, destination)
		err := `{ echo "Couldn't determine target" && exit 1; }`
		cmd = append(cmd, fmt.Sprintf("{ %s; }", strings.Join([]string{ifFile, ifInDir, ifDir, err}, " || ")))
	}

	if dc.Permissions != 0 {
		ifFile := fmt.Sprintf("{ if [ -f %s ]; then chmod %o %s; fi; }", destination, dc.Permissions, destination)
		ifInDir := fmt.Sprintf("{ if [ -d %s && -f %s/%s ]; then chmod %o %s/%s; fi; }", destination, destination,
			filename, dc.Permissions, destination, filename)
		ifDir := fmt.Sprintf("{ if [ -d %s ]; then chmod %o %s; fi; }", destination, dc.Permissions, destination)
		err := `{ echo "Couldn't determine target" && exit 1; }`
		cmd = append(cmd, fmt.Sprintf("{ %s; }", strings.Join([]string{ifFile, ifInDir, ifDir, err}, " || ")))
	}

	return strings.Join(cmd, " && ")
}

func (dc *DownloadCommand) Logging() string {
	sList := []string{"[DWNLOAD]"}

	if dc.Owner != "" && dc.Owner != "root" {
		sList = append(sList, fmt.Sprintf("[CHOWN:%s]", dc.Owner))
	}

	if dc.Permissions != 0 {
		sList = append(sList, fmt.Sprintf("[CHMOD:%.4o]", dc.Permissions))
	}

	sList = append(sList, fmt.Sprintf(" >> downloading file %q", dc.Url))
	if dc.Extract {
		sList = append(sList, " and extracting archive")
	}
	if dc.Destination != "" {
		sList = append(sList, fmt.Sprintf(" to %q", dc.Destination))
	}
	return strings.Join(sList, "")
}
