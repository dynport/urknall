package cmd

func ExampleAnd() {
	And(UpdatePackages(), InstallPackages("tmux", "vim", "zsh"))
}

func ExampleAsUser() {
	AsUser("gfrey", `echo "hello"`)
	AsUser("gfrey", Mkdir("${HOME}/.profile.d", "gfrey", 0))
}

func ExampleDownloadAndExtract() {
	DownloadAndExtract("http://code.google.com/p/go/downloads/detail?name=go1.2.linux-amd64.tar.gz&can=2&q=A", "/usr/local/go")
}

func ExampleDownloadToFile() {
	// Creates file "/usr/local/bin/docker-0.7.0"
	DownloadToFile("http://get.docker.io/builds/Linux/x86_64/docker-0.7.0", "/usr/local/bin", "root", 0700)
	// Creates file "/usr/local/bin/docker"
	DownloadToFile("http://get.docker.io/builds/Linux/x86_64/docker-0.7.0", "/usr/local/bin/docker", "root", 0700)
}

func ExampleIf() {
	// If the file "/tmp/foo" exists, remove it.
	If("-f /tmp/foo", "rm -f /tmp/foo")
}

func ExampleIfNot() {
	// If the directory "/tmp/foo" does not exist, create it and create a file in it.
	IfNot("-d /tmp/foo", And("mkdir -p /tmp/foo", "touch /tmp/foo/bar"))
}

func ExampleInstallPackages() {
	// Install some packages.
	InstallPackages("vim", "tmux", "zsh")
}

func ExampleMkdir() {
	// Create directory "/tmp/foo" with owner "root" and default permissions.
	Mkdir("/tmp/foo", "", 0)
	// Create directory "/tmp/bar" with owner "gfrey" and some obscure permissions.
	Mkdir("/tmp/bar", "gfrey", 0752)
}

func ExampleOr() {
	// Test whether platform is precise or raring, if none of those print an error and exit.
	// Given that "raring" is the current platform, the first command would fail (exit code not equal to 0), so the next
	// command would be run. The second command succeeds and therefore the remaining command is skipped, i.e. not
	// executed. The overall exit code would be 0 in this example.
	Or(
		"lsb_release -c | grep precise",
		"lsb_release -c | grep raring",
		And(`echo "no supported plattform found"`, "exit 1"))
}

func ExampleWaitForFile() {
	// Will wait at most 10 seconds for the file "/tmp/foo". Exit code is 1 in case of an error.
	WaitForFile("/tmp/foo", 10)
}

func ExampleWaitForUnixSocket() {
	// Will wait at most 15 seconds for the socket "/tmp/foo.sock". Exit code is 1 in case of an error.
	WaitForUnixSocket("/tmp/foo.sock", 15)
}
