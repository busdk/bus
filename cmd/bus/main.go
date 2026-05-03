package main

import (
	"os"

	"bus/internal/dispatch"
)

func main() {
	if handled, code := handleMetadataHelp(os.Args[1:], os.Stdout, os.Stderr); handled {
		os.Exit(code)
	}
	os.Exit(dispatch.Run(os.Args, os.Environ(), os.Stdin, os.Stdout, os.Stderr))
}
