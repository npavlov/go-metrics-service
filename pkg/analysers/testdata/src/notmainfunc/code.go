package main

import "os"

// No error should be reported here because there is no Exit in main function
func fooo() {
	os.Exit(1)
}
