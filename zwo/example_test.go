package zwo

import (
	. "github.com/dynport/zwo/cmd"
)

// Simple example to add two commands to a runlist.
func Example_Execute() {
	rl := &Runlist{} // Create the runlist.
	// Restart the nginx service.
	rl.Add("service nginx restart")
	// Install some packages using a helper from the cmd package.
	rl.Add(InstallPackages("curl", "vim"))
}

// Simple example to run a command as user 'nobody'.
func Example_ExecuteAsUser() {
	rl := &Runlist{} // Create the runlist.
	// As user 'nobody' configure, build and install some program from source to the private folder '~/usr'.
	rl.Add(
		AsUser("nobody",
			And("configure --prefix=${HOME}/usr",
				"make",
				"make install")))
}

// Simple example to add two action to a runlist that will create files.
func Example_File() {
	rl := &Runlist{} // Create the runlist.
	// Add a file /tmp/bar with content "foo" with owner set to "root" and mode set to 0600.
	rl.Add(File("/tmp/bar", "foo", "root", 0600))
	// Add a file /tmp/foo with content "bar" with owner set to "nobody" and mode set to 0666.
	rl.Add(File("/tmp/foo", "bar", "nobody", 0200))
}
