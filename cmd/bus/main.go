package main

import (
	"os"

	"bus/internal/dispatch"
)

func main() {
	os.Exit(dispatch.Run(os.Args, os.Environ(), os.Stdin, os.Stdout, os.Stderr))
}
