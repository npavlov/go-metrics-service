package main

import "os"

func main() {
	os.Exit(1) // want "calling os.Exit in function main"
}
