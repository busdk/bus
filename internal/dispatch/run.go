package dispatch

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// Run dispatches to a "bus-<command>" executable located on PATH.
func Run(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: bus <command> [args...]")
		return 2
	}

	subcommand := args[1]
	executable := "bus-" + subcommand

	path, err := exec.LookPath(executable)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"bus: subcommand %q not found; expected executable named %s in PATH\n",
			subcommand,
			executable,
		)
		return 127
	}

	cmd := exec.Command(path, args[2:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if code := exitErr.ExitCode(); code >= 0 {
				return code
			}
		}
		fmt.Fprintln(stderr, "bus: "+err.Error())
		return 1
	}

	return 0
}
