package dispatch

import (
	"bufio"
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

const version = "dev"

// Run dispatches to a "bus-<command>" executable located on PATH.
func Run(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	busfileOpts, busfileMode, err := parseBusfileMode(args[1:])
	if err != nil {
		fmt.Fprintf(stderr, "bus: invalid usage: %v\n", err)
		return 2
	}
	if busfileMode {
		return runBusfiles(busfileOpts, env, stdin, stdout, stderr)
	}

	parsed, err := parseGlobalFlags(args[1:])
	if err != nil {
		fmt.Fprintf(stderr, "bus: invalid usage: %v\n", err)
		writeUsage(env, stderr)
		return 2
	}
	if parsed.help {
		writeHelp(env, stdout)
		return 0
	}
	if parsed.version {
		fmt.Fprintf(stdout, "bus %s\n", version)
		return 0
	}
	if parsed.subcommand == "" {
		writeUsage(env, stderr)
		return 2
	}

	subcommand := parsed.subcommand
	executable := "bus-" + subcommand

	path, err := lookPathEnv(executable, env)
	if err != nil {
		if subcommand == "help" {
			writeUsage(env, stderr)
			return 2
		}
		fmt.Fprintf(stderr, "bus: missing subcommand: %s; expected executable named %s in PATH\n", subcommand, executable)
		writeUsage(env, stderr)
		return 127
	}

	childArgs := append([]string{}, parsed.passThroughFlags...)
	childArgs = append(childArgs, parsed.subcommandArgs...)
	cmd := exec.Command(path, childArgs...)
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

type busfileOptions struct {
	check       bool
	trace       bool
	transaction string
	scope       string
	files       []string
}

type busfileCommand struct {
	File string
	Line int
	Raw  string
	Argv []string
}

type busfileError struct {
	File     string
	Line     int
	Message  string
	ExitCode int
}

func (e busfileError) Error() string {
	if e.File != "" && e.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
	}
	if e.File != "" {
		return fmt.Sprintf("%s: %s", e.File, e.Message)
	}
	return e.Message
}

func parseBusfileMode(args []string) (busfileOptions, bool, error) {
	opts := busfileOptions{
		transaction: "none",
		scope:       "file",
	}
	if len(args) == 0 {
		return opts, false, nil
	}

	sawBusfileFlag := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			return opts, false, nil
		case arg == "--check":
			opts.check = true
			sawBusfileFlag = true
		case arg == "--trace":
			opts.trace = true
			sawBusfileFlag = true
		case strings.HasPrefix(arg, "--transaction="):
			opts.transaction = strings.TrimPrefix(arg, "--transaction=")
			sawBusfileFlag = true
		case arg == "--transaction":
			if i+1 >= len(args) {
				return opts, true, fmt.Errorf("missing value for --transaction")
			}
			opts.transaction = args[i+1]
			sawBusfileFlag = true
			i++
		case strings.HasPrefix(arg, "--scope="):
			opts.scope = strings.TrimPrefix(arg, "--scope=")
			sawBusfileFlag = true
		case arg == "--scope":
			if i+1 >= len(args) {
				return opts, true, fmt.Errorf("missing value for --scope")
			}
			opts.scope = args[i+1]
			sawBusfileFlag = true
			i++
		case strings.HasPrefix(arg, "-"):
			if len(opts.files) > 0 || sawBusfileFlag {
				return opts, true, fmt.Errorf("unknown busfile option %s", arg)
			}
			return opts, false, nil
		default:
			if len(opts.files) == 0 && !isBusfilePath(arg) {
				return opts, false, nil
			}
			if len(opts.files) > 0 && !isBusfilePath(arg) {
				return opts, true, fmt.Errorf("expected busfile path, got %q", arg)
			}
			opts.files = append(opts.files, arg)
		}
	}

	if len(opts.files) == 0 {
		if sawBusfileFlag {
			return opts, true, fmt.Errorf("missing busfile path")
		}
		return opts, false, nil
	}
	if !isValidScope(opts.scope) {
		return opts, true, fmt.Errorf("invalid --scope %q", opts.scope)
	}
	if !isValidTransaction(opts.transaction) {
		return opts, true, fmt.Errorf("invalid --transaction %q", opts.transaction)
	}
	return opts, true, nil
}

func runBusfiles(opts busfileOptions, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if opts.transaction != "none" {
		fmt.Fprintf(stderr, "bus: invalid usage: transaction provider %q is not implemented\n", opts.transaction)
		return 2
	}

	commands, err := preflightBusfiles(opts.files)
	if err != nil {
		var bfErr busfileError
		if errors.As(err, &bfErr) {
			fmt.Fprintln(stderr, bfErr.Error())
			return bfErr.ExitCode
		}
		fmt.Fprintf(stderr, "bus: %v\n", err)
		return 1
	}

	if opts.check {
		return 0
	}

	for _, command := range commands {
		if opts.trace {
			fmt.Fprintf(stdout, "%s:%d: bus %s\n", command.File, command.Line, strings.Join(command.Argv, " "))
		}
		commandEnv := withBusfileEnv(env, command.File, command.Line)
		code, runErr := runBusfileCommand(command, commandEnv, stdin, stdout, stderr)
		if runErr != nil {
			fmt.Fprintf(stderr, "%s:%d: %v\n", command.File, command.Line, runErr)
			return code
		}
	}
	return 0
}

func preflightBusfiles(files []string) ([]busfileCommand, error) {
	var commands []busfileCommand
	stack := map[string]bool{}
	for _, path := range files {
		if err := collectBusfileCommands(path, stack, &commands); err != nil {
			return nil, err
		}
	}
	return commands, nil
}

func collectBusfileCommands(path string, stack map[string]bool, commands *[]busfileCommand) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	if stack[absPath] {
		return busfileError{File: path, Message: "syntax error: include cycle detected", ExitCode: 65}
	}
	stack[absPath] = true
	defer delete(stack, absPath)

	file, err := os.Open(path)
	if err != nil {
		return busfileError{File: path, Message: err.Error(), ExitCode: 2}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		argv, parseErr := tokenizeBusLine(trimmed)
		if parseErr != nil {
			return busfileError{
				File:     path,
				Line:     lineNo,
				Message:  "syntax error: " + parseErr.Error(),
				ExitCode: 65,
			}
		}
		if len(argv) == 1 && strings.HasSuffix(argv[0], ".bus") {
			includePath := argv[0]
			if !filepath.IsAbs(includePath) {
				includePath = filepath.Join(filepath.Dir(path), includePath)
			}
			if err := collectBusfileCommands(includePath, stack, commands); err != nil {
				return err
			}
			continue
		}
		if len(argv) == 0 {
			return busfileError{
				File:     path,
				Line:     lineNo,
				Message:  "syntax error: empty command",
				ExitCode: 65,
			}
		}
		*commands = append(*commands, busfileCommand{
			File: path,
			Line: lineNo,
			Raw:  trimmed,
			Argv: argv,
		})
	}
	if err := scanner.Err(); err != nil {
		return busfileError{File: path, Message: err.Error(), ExitCode: 2}
	}
	return nil
}

func tokenizeBusLine(line string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false
	tokenStarted := false

	flush := func() {
		tokens = append(tokens, current.String())
		current.Reset()
		tokenStarted = false
	}

	for _, r := range line {
		if escaped {
			current.WriteRune(r)
			escaped = false
			tokenStarted = true
			continue
		}
		if inSingle {
			if r == '\'' {
				inSingle = false
				continue
			}
			current.WriteRune(r)
			tokenStarted = true
			continue
		}
		if inDouble {
			if r == '"' {
				inDouble = false
				continue
			}
			current.WriteRune(r)
			tokenStarted = true
			continue
		}

		switch r {
		case '\\':
			escaped = true
			tokenStarted = true
		case '\'':
			inSingle = true
			tokenStarted = true
		case '"':
			inDouble = true
			tokenStarted = true
		case ' ', '\t':
			if tokenStarted {
				flush()
			}
		case '|', ';', '<', '>':
			return nil, fmt.Errorf("disallowed token %q", string(r))
		default:
			current.WriteRune(r)
			tokenStarted = true
		}
	}

	if escaped {
		return nil, fmt.Errorf("invalid trailing escape")
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote")
	}
	if tokenStarted {
		flush()
	}
	return tokens, nil
}

func runBusfileCommand(command busfileCommand, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if len(command.Argv) == 0 {
		return 65, fmt.Errorf("syntax error: empty command")
	}
	subcommand := command.Argv[0]
	executable := "bus-" + subcommand

	path, err := lookPathEnv(executable, env)
	if err != nil {
		return 127, fmt.Errorf("dispatch error: unknown target %q", subcommand)
	}

	cmd := exec.Command(path, command.Argv[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if code := exitErr.ExitCode(); code >= 0 {
				return code, fmt.Errorf("command failed (exit %d): %s", code, command.Raw)
			}
		}
		return 1, fmt.Errorf("dispatch error: %v", err)
	}

	return 0, nil
}

func withBusfileEnv(env []string, file string, line int) []string {
	withBatch := upsertEnv(env, "BUS_BATCH", "1")
	withFile := upsertEnv(withBatch, "BUS_BUSFILE", file)
	return upsertEnv(withFile, "BUS_BUSFILE_LINE", fmt.Sprintf("%d", line))
}

func upsertEnv(env []string, key, value string) []string {
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

func isBusfilePath(path string) bool {
	if strings.HasSuffix(path, ".bus") {
		return true
	}
	if !candidateExists(path) {
		return false
	}
	line, err := readFirstLine(path)
	if err != nil {
		return false
	}
	return line == "#!/usr/bin/bus" || line == "#!/usr/bin/env bus"
}

func readFirstLine(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func isValidTransaction(value string) bool {
	switch value {
	case "none", "git", "snapshot", "copy":
		return true
	default:
		return false
	}
}

func isValidScope(value string) bool {
	switch value {
	case "file", "batch":
		return true
	default:
		return false
	}
}

type parseResult struct {
	help             bool
	version          bool
	subcommand       string
	passThroughFlags []string
	subcommandArgs   []string
}

func parseGlobalFlags(args []string) (parseResult, error) {
	parsed := parseResult{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if i+1 >= len(args) {
				break
			}
			parsed.subcommand = args[i+1]
			parsed.subcommandArgs = append([]string{}, args[i+2:]...)
			return parsed, nil
		}
		switch {
		case arg == "-h" || arg == "--help":
			parsed.help = true
			return parsed, nil
		case arg == "-V" || arg == "--version":
			parsed.version = true
			return parsed, nil
		case arg == "-q" || arg == "--quiet" || arg == "--no-color":
			parsed.passThroughFlags = append(parsed.passThroughFlags, arg)
		case arg == "-v" || arg == "--verbose":
			parsed.passThroughFlags = append(parsed.passThroughFlags, arg)
		case strings.HasPrefix(arg, "-") && len(arg) > 2 && strings.Trim(arg[1:], "v") == "":
			for range len(arg[1:]) {
				parsed.passThroughFlags = append(parsed.passThroughFlags, "-v")
			}
		case strings.HasPrefix(arg, "--color="):
			parsed.passThroughFlags = append(parsed.passThroughFlags, arg)
		case strings.HasPrefix(arg, "--format="):
			parsed.passThroughFlags = append(parsed.passThroughFlags, arg)
		case arg == "-C" || arg == "--chdir" || arg == "-o" || arg == "--output" || arg == "-f" || arg == "--format" || arg == "--color":
			if i+1 >= len(args) {
				return parsed, fmt.Errorf("missing value for %s", arg)
			}
			parsed.passThroughFlags = append(parsed.passThroughFlags, arg, args[i+1])
			i++
		case strings.HasPrefix(arg, "-"):
			return parsed, fmt.Errorf("unknown flag %s", arg)
		default:
			parsed.subcommand = arg
			parsed.subcommandArgs = append([]string{}, args[i+1:]...)
			return parsed, nil
		}
	}
	return parsed, nil
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

func writeHelp(env []string, stdout io.Writer) {
	fmt.Fprintln(stdout, "usage: bus [global-flags] <command> [args...]")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Global flags:")
	fmt.Fprintln(stdout, "  -h, --help")
	fmt.Fprintln(stdout, "  -V, --version")
	fmt.Fprintln(stdout, "  -v, --verbose")
	fmt.Fprintln(stdout, "  -q, --quiet")
	fmt.Fprintln(stdout, "  -C, --chdir <dir>")
	fmt.Fprintln(stdout, "  -o, --output <file>")
	fmt.Fprintln(stdout, "  -f, --format <format>")
	fmt.Fprintln(stdout, "  --color <auto|always|never>")
	fmt.Fprintln(stdout, "  --no-color")
	fmt.Fprintln(stdout, "  --")
	subcommands := listSubcommands(env)
	if len(subcommands) == 0 {
		return
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "available commands:")
	for _, name := range subcommands {
		fmt.Fprintf(stdout, "  %s\n", name)
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
