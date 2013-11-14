package zwo

import (
	"fmt"
)

// The example given would create the "/tmp/foo" folder.
func ExampleExecute() {
	Execute("mkdir -p /tmp/foo")
}

// The example given would install the basic packages for building stuff from source ("curl", "build-essential" and
// "git-core").
func ExampleInstallPackages() {
	InstallPackages("curl", "build-essential", "git-core")
}

// The given example will download the godeb archive to the "/tmp" folder, extract it and remove the downloaded
// archive.
func ExampleAnd() {
	url := "https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz"

	download := fmt.Sprintf("curl -SsfLO %s", url)
	extract := "tar xvfz godeb-amd64.tar.gz"

	And(Execute("cd /tmp"),
		Execute(download),
		Execute(extract),
		Execute("rm -f /tmp/godeb-amd64.tar.gz"))
}

//  The given example will download the godeb file, but only try to extract if the downloaded archive exists and is a
//  regular file.
func ExampleIf() {
	url := "https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz"

	download := fmt.Sprintf("curl -SsfLO %s", url)
	extract := "tar xvfz godeb-amd64.tar.gz"

	And(Execute("cd /tmp"),
		Execute(download))
	If("-f /tmp/godeb-amd64.tar.gz",
		And(Execute(extract),
			Execute("rm -f /tmp/godeb-amd64.tar.gz")))
}

// This example will download and extract the godeb binary only if it doesn't yet exist (note this makes the command
// idempotent, what is not a requirement for zwo).
func ExampleIfNot() {
	url := "https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz"

	download := fmt.Sprintf("curl -SsfLO %s", url)
	extract := "tar xvfz godeb-amd64.tar.gz"

	IfNot("-f /tmp/godeb",
		And(Execute("cd /tmp"),
			Execute(download),
			Execute(extract),
			Execute("rm -f /tmp/godeb-amd64.tar.gz")))
}

// Simple example to add two commands to a runlist. Please note that usually you don't need to create the runlist, as
// it is given to the Compile method of the Package interface you should implement for your packages.
func ExampleRunlist_AddCommands() {
	rl := &Runlist{} // Create the runlist.
	rl.AddCommands(
		// Install some packages.
		InstallPackages("curl", "vim"),
		// Create a file /tmp/bar containing "foo".
		Execute("echo 'foo' > /tmp/bar"))
}

// Simple example to add two action to a runlist that will create files. Please note that usually you don't need to
// create the runlist, as it is given to the Compile method of the Package interface you should implement for your
// packages.
func ExampleRunlist_AddFiles() {
	rl := &Runlist{} // Create the runlist.
	rl.AddFiles(
		// Add a file /tmp/bar with content "foo" with owner set to "root" and mode set to 0600.
		WriteFile("/tmp/bar", "foo", "root", 0600),
		// Add a file /tmp/foo with content "bar" with owner set to "nobody" and mode set to 0666.
		WriteFile("/tmp/foo", "bar", "nobody", 0200))
}
