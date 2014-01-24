package cmd

func ExampleWriteFile() {
	// Create a file "/tmp/foo" with the content "Hello, World!". No owner and permissions will be set.
	WriteFile("/tmp/foo", "Hello, World!", "", 0)
	// Create a file "/tmp/bar" with the content "Hello, World!". Set owner to "gfrey" and permissions 0600.
	WriteFile("/tmp/foo", "Hello, World!", "gfrey", 0600)
}
