package urknall

import "os"

func ExampleOpenLogger() {
	logger := OpenLogger(os.Stdout)
	defer logger.Close()

	// short: defer OpenLogger(os.Stdout).Close()
}
