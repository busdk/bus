package dispatch_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"bus/internal/dispatch"
)

func TestRunNoArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus"}, os.Environ(), nil, io.Discard, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "usage: bus <command> [args...]") {
		t.Fatalf("expected usage message, got %q", stderr.String())
	}
}

func TestRunMissingSubcommand(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	env := prependPath(os.Environ(), tempDir)

	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "missing"}, env, nil, io.Discard, &stderr)

	if code != 127 {
		t.Fatalf("expected exit code 127, got %d", code)
	}
	expected := `bus: subcommand "missing" not found; expected executable named bus-missing in PATH`
	if !strings.Contains(stderr.String(), expected) {
		t.Fatalf("expected error %q, got %q", expected, stderr.String())
	}
}

func TestRunDispatchesAndPassesArgs(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "PRIMARY")
	env := prependPath(os.Environ(), tempDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "accounts", "foo", "bar"}, env, nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.String() != "PRIMARY:foo bar\n" {
		t.Fatalf("expected stdout %q, got %q", "PRIMARY:foo bar\n", stdout.String())
	}
}

func TestRunPassesExitCode(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "EXITCODE")
	env := prependPath(os.Environ(), tempDir)
	env = setEnv(env, "BUS_SUBCMD_EXIT_CODE", "7")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "accounts"}, env, nil, &stdout, &stderr)

	if code != 7 {
		t.Fatalf("expected exit code 7, got %d", code)
	}
}

func TestRunUsesFrontOfPath(t *testing.T) {
	t.Parallel()

	firstDir := t.TempDir()
	secondDir := t.TempDir()

	buildFakeSubcommand(t, firstDir, "FIRST")
	buildFakeSubcommand(t, secondDir, "SECOND")

	env := prependPath(os.Environ(), secondDir)
	env = prependPath(env, firstDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "accounts", "ok"}, env, nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.String() != "FIRST:ok\n" {
		t.Fatalf("expected stdout %q, got %q", "FIRST:ok\n", stdout.String())
	}
}

func buildFakeSubcommand(t *testing.T, targetDir, label string) string {
	t.Helper()

	sourceDir := t.TempDir()
	source := fmt.Sprintf(`package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const prefix = %q

func main() {
	args := strings.Join(os.Args[1:], " ")
	fmt.Printf("%%s:%%s\n", prefix, args)

	if msg := os.Getenv("BUS_SUBCMD_STDERR"); msg != "" {
		fmt.Fprintln(os.Stderr, msg)
	}

	if codeText := os.Getenv("BUS_SUBCMD_EXIT_CODE"); codeText != "" {
		if code, err := strconv.Atoi(codeText); err == nil {
			os.Exit(code)
		}
	}
}
`, label)

	sourcePath := filepath.Join(sourceDir, "main.go")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatalf("write fake main.go: %v", err)
	}

	outputName := "bus-accounts"
	if runtime.GOOS == "windows" {
		outputName += ".exe"
	}
	outputPath := filepath.Join(targetDir, outputName)

	cmd := exec.Command("go", "build", "-o", outputPath)
	cmd.Dir = sourceDir
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	return outputPath
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	updated := make([]string, 0, len(env)+1)
	found := false
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			updated = append(updated, prefix+value)
			found = true
			continue
		}
		updated = append(updated, entry)
	}
	if !found {
		updated = append(updated, prefix+value)
	}
	return updated
}

func prependPath(env []string, dir string) []string {
	pathValue, _ := lookupEnv(env, "PATH")
	if pathValue == "" {
		return setEnv(env, "PATH", dir)
	}
	return setEnv(env, "PATH", dir+string(os.PathListSeparator)+pathValue)
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
