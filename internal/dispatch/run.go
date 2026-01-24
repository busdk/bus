package dispatch

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Run dispatches to a "bus-<command>" executable located on PATH.
func Run(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 2 {
		writeUsage(env, stderr)
		return 2
	}

	subcommand := args[1]
	executable := "bus-" + subcommand

	path, err := lookPathEnv(executable, env)
	if err != nil {
		fmt.Fprintf(stderr, "bus: missing subcommand: %s; expected executable named %s in PATH\n", subcommand, executable)
		writeUsage(env, stderr)
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

func writeUsage(env []string, stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: bus <command> [args...]")
	subcommands := listSubcommands(env)
	if len(subcommands) == 0 {
		return
	}
	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr, "available commands:")
	for _, name := range subcommands {
		fmt.Fprintf(stderr, "  %s\n", name)
	}
}

// Issue: https://github.com/busdk/bus/issues/2 - enumerate bus-* executables on PATH.
func listSubcommands(env []string) []string {
	pathValue, _ := lookupEnv(env, "PATH")
	if pathValue == "" {
		return nil
	}

	exts := map[string]struct{}{}
	if runtime.GOOS == "windows" {
		for _, ext := range windowsPathExts(env) {
			if ext == "" {
				continue
			}
			exts[strings.ToLower(ext)] = struct{}{}
		}
	}

	seen := map[string]struct{}{}
	for _, dir := range filepath.SplitList(pathValue) {
		if dir == "" {
			dir = "."
		}
		func() {
			dirHandle, err := os.Open(dir)
			if err != nil {
				return
			}
			defer dirHandle.Close()

			for {
				entries, err := dirHandle.ReadDir(128)
				if len(entries) == 0 && err == nil {
					break
				}
				for _, entry := range entries {
					name := entry.Name()
					if !strings.HasPrefix(name, "bus-") {
						continue
					}
					fullPath := filepath.Join(dir, name)
					if !candidateExists(fullPath) {
						continue
					}
					command, ok := subcommandFromFile(name, exts)
					if !ok {
						continue
					}
					if _, ok := seen[command]; ok {
						continue
					}
					seen[command] = struct{}{}
				}
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					break
				}
			}
		}()
	}

	subcommands := make([]string, 0, len(seen))
	for name := range seen {
		subcommands = append(subcommands, name)
	}
	sort.Strings(subcommands)
	return subcommands
}

func subcommandFromFile(name string, windowsExts map[string]struct{}) (string, bool) {
	if !strings.HasPrefix(name, "bus-") {
		return "", false
	}

	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(name))
		if ext == "" {
			return "", false
		}
		if _, ok := windowsExts[ext]; !ok {
			return "", false
		}
		name = strings.TrimSuffix(name, ext)
	}

	command := strings.TrimPrefix(name, "bus-")
	if command == "" {
		return "", false
	}
	return command, true
}

func lookPathEnv(file string, env []string) (string, error) {
	pathValue, _ := lookupEnv(env, "PATH")
	if pathValue == "" {
		return "", exec.ErrNotFound
	}

	if hasPathSeparator(file) {
		if candidateExists(file) {
			return file, nil
		}
		return "", exec.ErrNotFound
	}

	for _, dir := range filepath.SplitList(pathValue) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, file)
		if found, ok := resolveCandidate(candidate, env); ok {
			return found, nil
		}
	}

	return "", exec.ErrNotFound
}

func resolveCandidate(path string, env []string) (string, bool) {
	if runtime.GOOS != "windows" {
		return path, candidateExists(path)
	}

	if filepath.Ext(path) != "" {
		return path, candidateExists(path)
	}

	exts := windowsPathExts(env)
	for _, ext := range exts {
		if ext == "" {
			continue
		}
		candidate := path + ext
		if candidateExists(candidate) {
			return candidate, true
		}
	}

	return "", false
}

func windowsPathExts(env []string) []string {
	value, _ := lookupEnv(env, "PATHEXT")
	if value == "" {
		value = ".com;.exe;.bat;.cmd"
	}
	parts := strings.Split(strings.ToLower(value), ";")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if !strings.HasPrefix(part, ".") {
			parts[i] = "." + part
		}
	}
	return parts
}

func candidateExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

func hasPathSeparator(path string) bool {
	if strings.ContainsRune(path, filepath.Separator) {
		return true
	}
	if runtime.GOOS == "windows" && strings.Contains(path, "/") {
		return true
	}
	return false
}

func lookupEnv(env []string, key string) (string, bool) {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix), true
		}
	}
	return "", false
}
