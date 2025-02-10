package main

import "os"

func main() {
	os.Exit(1) // want "error: calling os.Exit in main function of main package"
}
