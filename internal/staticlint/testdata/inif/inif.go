package main

import "os"

func main() {
	if len(os.Args) > 1 {
		os.Exit(1) // want "call os.Exit at main func of package main"
	}
}
