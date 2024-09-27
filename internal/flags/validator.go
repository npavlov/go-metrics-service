package flags

import (
	"flag"
	"fmt"
	"os"
)

func VerifyFlags() {
	// Check for unexpected flags or arguments
	if len(flag.Args()) > 0 {
		fmt.Printf("Error: Unexpected argument(s): %v\n", flag.Args())
		flag.Usage() // Optionally print usage
		os.Exit(1)   // Exit with an error code
	}

}
