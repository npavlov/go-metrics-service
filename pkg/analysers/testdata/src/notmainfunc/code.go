package main

import "os"

func fooo() {
	os.Exit(1) // want "warning: calling os.Exit"
}
