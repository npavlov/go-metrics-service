package test

import "fmt"

// No error should be reported here because there is no os.Exit
func main() {
	var a = 1
	fmt.Println(a)
}
