package urknall

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/dynport/urknall/cmd"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Generate a Dockerfile from the commands collected on the runlist.
func (rl *Runlist) DockerArchive(from string, w io.Writer) (e error) {
	tw := tar.NewWriter(w)
	defer tw.Close()

	dfLines := []string{
		"FROM " + from,
	}

	for _, c := range rl.commands {
		var sCmd string
		switch t := c.(type) {
		case cmd.DockerCommand:
			sCmd = t.Docker()
		case *cmd.FileCommand:
			content := bytes.NewBufferString(t.Content)
			srcName := filepath.Join(rl.name, "files", path.Base(t.Path))
			if e = writeFileToTarArchive(tw, content, srcName); e != nil {
				return e
			}

			sCmd = fmt.Sprintf("ADD %s %s", srcName, t.Path)
		default:
			sCmd = fmt.Sprintf("RUN %s", c.Shell())
		}
		dfLines = append(dfLines, sCmd)
	}

	content := bytes.NewBufferString(strings.Join(dfLines, "\n"))
	return writeFileToTarArchive(tw, content, filepath.Join(rl.name, "Dockerfile"))
}

func writeFileToTarArchive(tw *tar.Writer, buf *bytes.Buffer, filename string) (e error) {
	header := &tar.Header{
		Name:    filename,
		Size:    int64(buf.Len()),
		ModTime: time.Now(),
		Mode:    0644,
	}

	if e := tw.WriteHeader(header); e != nil {
		return e
	}

	_, e = io.Copy(tw, buf)
	return e
}
