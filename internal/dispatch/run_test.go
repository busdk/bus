package dispatch_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"testing/quick"

	"bus/internal/dispatch"
)

func TestRunNoArgs(t *testing.T) {
	t.Parallel()

	// Issue: https://github.com/busdk/bus/issues/2 - no-args help lists subcommands.
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "HELP")
	env := prependPath(os.Environ(), tempDir)

	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus"}, env, nil, io.Discard, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "usage: bus <command> [args...]") {
		t.Fatalf("expected usage message, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "available commands:") {
		t.Fatalf("expected subcommand list header, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "accounts") {
		t.Fatalf("expected accounts in help output, got %q", stderr.String())
	}
}

// Run properties (Issue: https://github.com/busdk/bus/issues/2):
// - With no subcommand, help is printed and exit code is 2 regardless of env noise.
// - The help output always includes a subcommand list header.

func TestRunMissingSubcommand(t *testing.T) {
	t.Parallel()

	// Issue: https://github.com/busdk/bus/issues/2 - missing subcommand remains explicit.
	tempDir := t.TempDir()
	env := prependPath(os.Environ(), tempDir)

	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "missing"}, env, nil, io.Discard, &stderr)

	if code != 127 {
		t.Fatalf("expected exit code 127, got %d", code)
	}
	expected := `bus: missing subcommand: missing; expected executable named bus-missing in PATH`
	if !strings.Contains(stderr.String(), expected) {
		t.Fatalf("expected error %q, got %q", expected, stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage: bus <command> [args...]") {
		t.Fatalf("expected usage after error, got %q", stderr.String())
	}
}

func TestRunHelpWithoutSubcommandBinary(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "HELP")
	env := setEnv(os.Environ(), "PATH", tempDir)

	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "help"}, env, nil, io.Discard, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	output := stderr.String()
	if !strings.Contains(output, "usage: bus <command> [args...]") {
		t.Fatalf("expected usage output, got %q", output)
	}
	if strings.Contains(output, "missing subcommand") {
		t.Fatalf("unexpected missing subcommand output, got %q", output)
	}
	if !strings.Contains(output, "available commands:") {
		t.Fatalf("expected available commands, got %q", output)
	}
	if !strings.Contains(output, "accounts") {
		t.Fatalf("expected discovered command in help output, got %q", output)
	}
}

func TestRunHelpDispatchesWhenBinaryExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "help", "HELPER")
	env := setEnv(os.Environ(), "PATH", tempDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "help", "accounts"}, env, nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.String() != "HELPER:accounts\n" {
		t.Fatalf("expected stdout %q, got %q", "HELPER:accounts\n", stdout.String())
	}
}

func TestRunGlobalHelpShortCircuits(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "HELP")
	env := setEnv(os.Environ(), "PATH", tempDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "-q", "--verbose", "--help"}, env, nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage: bus [global-flags] <command> [args...]") {
		t.Fatalf("expected global help usage, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "available commands:") {
		t.Fatalf("expected available commands in help, got %q", stdout.String())
	}
}

func TestRunGlobalVersionShortCircuits(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "--color", "rainbow", "--version"}, os.Environ(), nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}
	if stdout.String() != "bus dev\n" {
		t.Fatalf("expected version output %q, got %q", "bus dev\n", stdout.String())
	}
}

func TestRunParsesGlobalFlagsBeforeSubcommand(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "status", "PRIMARY")
	env := setEnv(os.Environ(), "PATH", tempDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "-q", "-C", "/", "--color=never", "status", "--version"}, env, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.String() != "PRIMARY:-q -C / --color=never --version\n" {
		t.Fatalf("unexpected delegated args: %q", stdout.String())
	}
}

func TestRunDoubleDashTerminatesGlobalParsing(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "status", "PRIMARY")
	env := setEnv(os.Environ(), "PATH", tempDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "--", "status", "--version"}, env, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.String() != "PRIMARY:--version\n" {
		t.Fatalf("unexpected delegated args with -- terminator: %q", stdout.String())
	}
}

func TestRunDispatchesAndPassesArgs(t *testing.T) {
	t.Parallel()

	// Issue: https://github.com/busdk/bus/issues/2 - dispatch behavior unchanged.
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "PRIMARY")
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

	// Issue: https://github.com/busdk/bus/issues/2 - exit codes pass through.
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "EXITCODE")
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

	// Issue: https://github.com/busdk/bus/issues/2 - PATH front wins.
	firstDir := t.TempDir()
	secondDir := t.TempDir()

	buildFakeSubcommand(t, firstDir, "accounts", "FIRST")
	buildFakeSubcommand(t, secondDir, "accounts", "SECOND")

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

func TestRunNoArgsShowsNoneWhenPathEmpty(t *testing.T) {
	t.Parallel()

	// Issue: https://github.com/busdk/bus/issues/2 - help shows empty list when no commands exist.
	env := setEnv(os.Environ(), "PATH", "")

	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus"}, env, nil, io.Discard, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if strings.Contains(stderr.String(), "available commands:") {
		t.Fatalf("expected no available commands section, got %q", stderr.String())
	}
}

func TestRunHelpMetamorphicPathOrder(t *testing.T) {
	t.Parallel()

	// Issue: https://github.com/busdk/bus/issues/2 - listing is stable across PATH order.
	firstDir := t.TempDir()
	secondDir := t.TempDir()

	buildFakeSubcommand(t, firstDir, "accounts", "FIRST")
	buildFakeSubcommand(t, secondDir, "ledger", "SECOND")

	envFirst := setEnv(os.Environ(), "PATH", firstDir+string(os.PathListSeparator)+secondDir)
	envSecond := setEnv(os.Environ(), "PATH", secondDir+string(os.PathListSeparator)+firstDir)

	firstList := helpSubcommands(t, envFirst)
	secondList := helpSubcommands(t, envSecond)

	if !reflect.DeepEqual(firstList, secondList) {
		t.Fatalf("expected identical command lists, got %v vs %v", firstList, secondList)
	}
}

func TestRunHelpMatchesReferenceListing(t *testing.T) {
	t.Parallel()

	// Issue: https://github.com/busdk/bus/issues/2 - help list matches reference scan.
	tempDir := t.TempDir()

	buildFakeSubcommand(t, tempDir, "accounts", "REF")
	buildFakeSubcommand(t, tempDir, "ledger", "REF")

	if runtime.GOOS != "windows" {
		nonExecPath := filepath.Join(tempDir, "bus-nonexec")
		if err := os.WriteFile(nonExecPath, []byte("skip"), 0o600); err != nil {
			t.Fatalf("write non-exec file: %v", err)
		}
	}

	env := setEnv(os.Environ(), "PATH", tempDir)
	got := helpSubcommands(t, env)
	want := referenceSubcommands(tempDir, env)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected help list %v, got %v", want, got)
	}
}

func TestRunNoArgsProperties(t *testing.T) {
	t.Parallel()

	// Issue: https://github.com/busdk/bus/issues/2 - no-args help is invariant over env noise.
	tempDir := t.TempDir()
	baseEnv := setEnv(os.Environ(), "PATH", tempDir)

	config := &quick.Config{
		MaxCount: 50,
		Rand:     rand.New(rand.NewSource(42)),
	}

	property := func(extras []string) bool {
		env := append([]string{}, baseEnv...)
		env = append(env, sanitizeEnvExtras(extras)...)

		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus"}, env, nil, io.Discard, &stderr)
		if code != 2 {
			return false
		}
		output := stderr.String()
		return strings.Contains(output, "usage: bus <command> [args...]") &&
			!strings.Contains(output, "available commands:")
	}

	if err := quick.Check(property, config); err != nil {
		t.Fatalf("property test failed: %v", err)
	}
}

func FuzzRunMissingSubcommand(f *testing.F) {
	// Issue: https://github.com/busdk/bus/issues/2 - missing subcommands return 127 without panicking.
	f.Add("accounts")
	f.Add("ledger")
	f.Add("foo/bar")

	f.Fuzz(func(t *testing.T, subcommand string) {
		tempDir := t.TempDir()
		env := setEnv(os.Environ(), "PATH", tempDir)

		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus", subcommand}, env, nil, io.Discard, &stderr)

		if code != 127 {
			t.Fatalf("expected exit code 127, got %d", code)
		}
		if !strings.Contains(stderr.String(), "subcommand") {
			t.Fatalf("expected subcommand error, got %q", stderr.String())
		}
	})
}

func helpSubcommands(t *testing.T, env []string) []string {
	t.Helper()

	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus"}, env, nil, io.Discard, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	return parseSubcommands(stderr.String())
}

func parseSubcommands(help string) []string {
	lines := strings.Split(help, "\n")
	var commands []string
	start := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "available commands:" {
			start = true
			continue
		}
		if !start {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		trimmed := strings.TrimSpace(line)
		commands = append(commands, trimmed)
	}
	return commands
}

func referenceSubcommands(dir string, env []string) []string {
	matches, err := filepath.Glob(filepath.Join(dir, "bus-*"))
	if err != nil {
		return nil
	}

	exts := parsePathExts(env)
	seen := map[string]struct{}{}
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil || info.IsDir() {
			continue
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
			continue
		}

		name := filepath.Base(match)
		if runtime.GOOS == "windows" {
			ext := strings.ToLower(filepath.Ext(name))
			if ext == "" {
				continue
			}
			if _, ok := exts[ext]; !ok {
				continue
			}
			name = strings.TrimSuffix(name, ext)
		}

		command := strings.TrimPrefix(name, "bus-")
		if command == "" {
			continue
		}
		seen[command] = struct{}{}
	}

	commands := make([]string, 0, len(seen))
	for command := range seen {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	return commands
}

func parsePathExts(env []string) map[string]struct{} {
	value, _ := lookupEnv(env, "PATHEXT")
	if value == "" {
		value = ".com;.exe;.bat;.cmd"
	}
	result := map[string]struct{}{}
	for _, part := range strings.Split(strings.ToLower(value), ";") {
		if part == "" {
			continue
		}
		if !strings.HasPrefix(part, ".") {
			part = "." + part
		}
		result[part] = struct{}{}
	}
	return result
}

func buildFakeSubcommand(t *testing.T, targetDir, subcommand, label string) string {
	t.Helper()

	sourceDir := t.TempDir()
	goMod := "module bus-subcmd\n\ngo 1.22\n"
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
	if err := os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatalf("write fake go.mod: %v", err)
	}

	outputName := "bus-" + subcommand
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

func sanitizeEnvExtras(extras []string) []string {
	cleaned := make([]string, 0, len(extras))
	for _, entry := range extras {
		if strings.HasPrefix(entry, "PATH=") || strings.HasPrefix(entry, "PATHEXT=") {
			continue
		}
		cleaned = append(cleaned, entry)
	}
	return cleaned
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

func TestRunBusfileExecutesCommands(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")
	buildFakeSubcommand(t, tempDir, "journal", "JOURNAL")

	busfile := filepath.Join(tempDir, "2024-02.bus")
	content := strings.Join([]string{
		"# month file",
		"accounts add --code 3000",
		"journal add --date 2024-02-29 --debit 1910=10.00 --credit 3000=10.00 --desc 'M2'",
		"",
	}, "\n")
	if err := os.WriteFile(busfile, []byte(content), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", busfile}, env, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "ACCOUNTS:add --code 3000\n") {
		t.Fatalf("expected accounts invocation, got %q", got)
	}
	if !strings.Contains(got, "JOURNAL:add --date 2024-02-29 --debit 1910=10.00 --credit 3000=10.00 --desc M2\n") {
		t.Fatalf("expected journal invocation, got %q", got)
	}
}

func TestRunBusfilePreflightStopsBeforeExecution(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	good := filepath.Join(tempDir, "2024-01.bus")
	bad := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(good, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write good busfile: %v", err)
	}
	if err := os.WriteFile(bad, []byte("accounts add --desc 'unterminated\n"), 0o600); err != nil {
		t.Fatalf("write bad busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", good, bad}, env, nil, &stdout, &stderr)
	if code != 65 {
		t.Fatalf("expected exit 65, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no command execution output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "syntax error: unterminated quote") {
		t.Fatalf("expected syntax error, got %q", stderr.String())
	}
}

func TestRunBusfileUnknownTargetPreflightStopsBeforeExecution(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	good := filepath.Join(tempDir, "2024-01.bus")
	bad := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(good, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write good busfile: %v", err)
	}
	if err := os.WriteFile(bad, []byte("bnak add --id 1\n"), 0o600); err != nil {
		t.Fatalf("write bad busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", good, bad}, env, nil, &stdout, &stderr)
	if code != 127 {
		t.Fatalf("expected exit 127, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no command execution output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "dispatch error: unknown target \"bnak\"") {
		t.Fatalf("expected unknown target preflight error, got %q", stderr.String())
	}
}

func TestRunBusfileIncludes(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	jan := filepath.Join(tempDir, "2024-01.bus")
	feb := filepath.Join(tempDir, "2024-02.bus")
	root := filepath.Join(tempDir, "all.bus")
	if err := os.WriteFile(jan, []byte("accounts add --code 1000\n"), 0o600); err != nil {
		t.Fatalf("write jan busfile: %v", err)
	}
	if err := os.WriteFile(feb, []byte("accounts add --code 2000\n"), 0o600); err != nil {
		t.Fatalf("write feb busfile: %v", err)
	}
	if err := os.WriteFile(root, []byte("2024-01.bus\n2024-02.bus\n"), 0o600); err != nil {
		t.Fatalf("write root busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", root}, env, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "ACCOUNTS:add --code 1000\n") || !strings.Contains(got, "ACCOUNTS:add --code 2000\n") {
		t.Fatalf("expected included files to execute, got %q", got)
	}
}

func TestRunBusfileCheckDoesNotExecute(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	file := filepath.Join(tempDir, "check.bus")
	if err := os.WriteFile(file, []byte("accounts add --code 3000\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "--check", file}, env, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no execution output in --check mode, got %q", stdout.String())
	}
}

func TestRunBusfileCheckFailsOnUnbalancedJournal(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "journal", "JOURNAL")

	file := filepath.Join(tempDir, "unbalanced.bus")
	content := "journal add --date 2024-02-29 --debit 1910=10.00 --credit 3000=9.99\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "--check", file}, env, nil, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no execution output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "validation error: journal add unbalanced entry") {
		t.Fatalf("expected unbalanced validation error, got %q", stderr.String())
	}
}

func TestRunBusfileCheckFailsOnInvalidBankValues(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "bank", "BANK")

	file := filepath.Join(tempDir, "bank-invalid.bus")
	content := "bank add transactions --set booked_date=2024-99-99 --set amount=NaN --set currency=EURO\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "--check", file}, env, nil, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no execution output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "validation error: bank add transactions invalid booked_date") {
		t.Fatalf("expected bank validation error, got %q", stderr.String())
	}
}

func TestRunBusfileApplyFailsBeforeExecutionOnValidationError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "journal", "JOURNAL")

	file := filepath.Join(tempDir, "apply-invalid.bus")
	content := "journal add --date 2024-02-29 --debit 1910=10.00 --credit 3000=9.99\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", file}, env, nil, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no execution output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "validation error: journal add unbalanced entry") {
		t.Fatalf("expected unbalanced validation error, got %q", stderr.String())
	}
}

func TestRunBusfileTrace(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	file := filepath.Join(tempDir, "trace.bus")
	if err := os.WriteFile(file, []byte("accounts add --code 3000\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := dispatch.Run([]string{"bus", "--trace", file}, env, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "trace.bus:1: bus accounts add --code 3000\n") {
		t.Fatalf("expected trace line, got %q", stdout.String())
	}
}

func TestRunBusfileTransactionFromDatapackageFallbackToNone(t *testing.T) {
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	busfile := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(busfile, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"file","fallback_to_none":true}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	withChdir(t, tempDir, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus", busfile}, env, nil, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "ACCOUNTS:list\n") {
			t.Fatalf("expected command execution, got %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "falling back to \"none\"") {
			t.Fatalf("expected fallback warning, got %q", stderr.String())
		}
	})
}

func TestRunBusfileTransactionFromDatapackageFailsWithoutFallback(t *testing.T) {
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	busfile := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(busfile, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"file","fallback_to_none":false}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	withChdir(t, tempDir, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus", busfile}, env, nil, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("expected exit 2, got %d (stderr: %q)", code, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected no command output, got %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "provider \"fs\" requires in-process tx runners") {
			t.Fatalf("expected fs unavailable error, got %q", stderr.String())
		}
	})
}

func TestRunBusfileCliTransactionOverrideFails(t *testing.T) {
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	busfile := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(busfile, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"none","scope":"file","fallback_to_none":true}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	withChdir(t, tempDir, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus", "--transaction", "fs", busfile}, env, nil, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("expected exit 2, got %d (stderr: %q)", code, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected no command output, got %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "provider \"fs\" requires in-process tx runners") {
			t.Fatalf("expected fs unavailable error, got %q", stderr.String())
		}
	})
}

func TestRunBusfilePreferencesOverrideDatapackage(t *testing.T) {
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	busfile := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(busfile, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"none","scope":"file","fallback_to_none":true}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	prefsPath := filepath.Join(tempDir, "prefs.json")
	prefs := `{"version":1,"values":{"bus.busfile.transaction.provider":"fs","bus.busfile.transaction.fallback_to_none":false}}`
	if err := os.WriteFile(prefsPath, []byte(prefs), 0o600); err != nil {
		t.Fatalf("write prefs: %v", err)
	}

	t.Setenv("BUS_PREFERENCES_PATH", prefsPath)
	env := prependPath(os.Environ(), tempDir)
	withChdir(t, tempDir, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus", busfile}, env, nil, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("expected exit 2, got %d (stderr: %q)", code, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected no command output, got %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "provider \"fs\" requires in-process tx runners") {
			t.Fatalf("expected fs unavailable error, got %q", stderr.String())
		}
	})
}

func TestRunBusfileShellLookupDisabledFailsWithoutInProcessRunner(t *testing.T) {
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	busfile := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(busfile, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}

	prefsPath := filepath.Join(tempDir, "prefs.json")
	prefs := `{"version":1,"values":{"bus.busfile.dispatch.shell_lookup_enabled":false}}`
	if err := os.WriteFile(prefsPath, []byte(prefs), 0o600); err != nil {
		t.Fatalf("write prefs: %v", err)
	}

	t.Setenv("BUS_PREFERENCES_PATH", prefsPath)
	env := prependPath(os.Environ(), tempDir)
	withChdir(t, tempDir, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus", busfile}, env, nil, &stdout, &stderr)
		if code != 127 {
			t.Fatalf("expected exit 127, got %d (stderr: %q)", code, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected no command output, got %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "shell lookup disabled and no in-process runner") {
			t.Fatalf("expected shell lookup disabled dispatch error, got %q", stderr.String())
		}
	})
}

func TestRunBusfileShellLookupDisabledViaDatapackage(t *testing.T) {
	tempDir := t.TempDir()
	buildFakeSubcommand(t, tempDir, "accounts", "ACCOUNTS")

	busfile := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(busfile, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"dispatch":{"shell_lookup_enabled":false}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	env := prependPath(os.Environ(), tempDir)
	withChdir(t, tempDir, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := dispatch.Run([]string{"bus", busfile}, env, nil, &stdout, &stderr)
		if code != 127 {
			t.Fatalf("expected exit 127, got %d (stderr: %q)", code, stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected no command output, got %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "shell lookup disabled and no in-process runner") {
			t.Fatalf("expected shell lookup disabled dispatch error, got %q", stderr.String())
		}
	})
}

func withChdir(t *testing.T, dir string, fn func()) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %q: %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(old)
	})
	fn()
}
