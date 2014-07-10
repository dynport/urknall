package main

import (
	"fmt"
	"path"
	"strings"
)

// Extract the file at the given directory. The following file extensions are currently supported (".tar", ".tgz",
// ".tar.gz", ".tbz", ".tar.bz2" for tar archives, and ".zip" for zipfiles).
func Extract(file, targetDir string) *ShellCommand {
	if targetDir == "" {
		panic("empty target directory given")
	}

	var extractCmd *ShellCommand
	switch {
	case strings.HasSuffix(file, ".tar"):
		extractCmd = extractTarArchive(file, targetDir, "")
	case strings.HasSuffix(file, ".tgz"):
		fallthrough
	case strings.HasSuffix(file, ".tar.gz"):
		extractCmd = extractTarArchive(file, targetDir, "gz")
	case strings.HasSuffix(file, ".tbz"):
		fallthrough
	case strings.HasSuffix(file, ".tar.bz2"):
		extractCmd = extractTarArchive(file, targetDir, "bz2")
	case strings.HasSuffix(file, ".zip"):
		extractCmd = &ShellCommand{Command: fmt.Sprintf("unzip -d %s %s", targetDir, file)}
	default:
		panic(fmt.Sprintf("type of file %q not a supported archive", path.Base(file)))
	}

	return And(
		Mkdir(targetDir, "", 0),
		extractCmd)
}

func extractTarArchive(path, targetDir, compression string) *ShellCommand {
	additionalCommand := ""
	switch compression {
	case "gz":
		additionalCommand = "z"
	case "bz2":
		additionalCommand = "j"
	}
	return And(
		fmt.Sprintf("cd %s", targetDir),
		fmt.Sprintf("tar xf%s %s", additionalCommand, path))
}
