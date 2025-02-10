package test

import "os"

func main() {
	os.Exit(1) // want "warning: calling os.Exit"
}
