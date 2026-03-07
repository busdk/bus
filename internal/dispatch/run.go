package dispatch

import (
	"bufio"
	"bus/internal/txfs"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
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
		if subcommand == "audit" {
			return runAuditAlias(parsed, env, stdin, stdout, stderr)
		}
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

func runAuditAlias(parsed parseResult, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(parsed.subcommandArgs) == 0 || parsed.subcommandArgs[0] != "evidence-coverage" {
		fmt.Fprintln(stderr, "bus: invalid usage: audit requires subcommand evidence-coverage")
		fmt.Fprintln(stderr, "usage: bus audit evidence-coverage [args...]")
		return 2
	}
	path, err := lookPathEnv("bus-validate", env)
	if err != nil {
		fmt.Fprintln(stderr, "bus: missing subcommand: audit evidence-coverage requires executable named bus-validate in PATH")
		return 127
	}
	childArgs := auditAliasArgs(parsed)
	cmd := exec.Command(path, childArgs...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if runErr := cmd.Run(); runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			if code := exitErr.ExitCode(); code >= 0 {
				return code
			}
		}
		fmt.Fprintln(stderr, "bus: "+runErr.Error())
		return 1
	}
	return 0
}

// auditAliasArgs maps `bus audit evidence-coverage ...` arguments to the bus-validate invocation.
// Used by: runAuditAlias dispatcher fallback when bus-audit is not present.
func auditAliasArgs(parsed parseResult) []string {
	aliasArgs := append([]string{}, parsed.subcommandArgs[1:]...)
	for _, arg := range aliasArgs {
		if arg == "-h" || arg == "--help" {
			childArgs := append([]string{}, parsed.passThroughFlags...)
			childArgs = append(childArgs, "--help", "evidence-coverage")
			return childArgs
		}
	}
	childArgs := append([]string{}, parsed.passThroughFlags...)
	childArgs = append(childArgs, "evidence-coverage")
	childArgs = append(childArgs, aliasArgs...)
	return childArgs
}

type busfileOptions struct {
	check          bool
	trace          bool
	transaction    string
	transactionSet bool
	scope          string
	scopeSet       bool
	files          []string
}

type busfileConfig struct {
	transactionProvider string
	transactionScope    string
	fallbackToNone      bool
	validationLevel     string
	shellLookupEnabled  bool
}

type busfileExecutor interface {
	Execute(command busfileCommand, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error)
}

type hybridBusfileExecutor struct {
	shellLookupEnabled bool
	resolvedTargets    map[string]string
}

func (e hybridBusfileExecutor) Execute(command busfileCommand, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if len(command.Argv) == 0 {
		return 65, fmt.Errorf("syntax error: empty command")
	}
	target := command.Argv[0]
	if runner, ok := inProcessModuleRunners[target]; ok {
		code, err := runner(command.Argv[1:], env, stdin, stdout, stderr)
		if err != nil {
			return 1, fmt.Errorf("dispatch error: %v", err)
		}
		if code != 0 {
			return code, fmt.Errorf("command failed (exit %d): %s", code, command.Raw)
		}
		return 0, nil
	}
	if !e.shellLookupEnabled {
		return 127, fmt.Errorf("dispatch error: no in-process runner for target %q and shell lookup is disabled", target)
	}
	return runBusfileCommand(command, env, stdin, stdout, stderr, e.resolvedTargets)
}

type inProcessModuleRunner func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error)
type inProcessTxModuleRunner func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, fs *txfs.FS) (int, error)

var inProcessModuleRunners = map[string]inProcessModuleRunner{}
var inProcessTxModuleRunners = map[string]inProcessTxModuleRunner{}

func runExternalBusModule(target string, args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	executable := "bus-" + target
	path, err := lookPathEnv(executable, env)
	if err != nil {
		return 127, fmt.Errorf("dispatch error: unknown target %q", target)
	}

	cmd := exec.Command(path, args...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

type fsTxJournal struct {
	State   string   `json:"state"`
	Scope   string   `json:"scope"`
	TxID    string   `json:"tx_id"`
	Files   []string `json:"files,omitempty"`
	Updated string   `json:"updated"`
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
			opts.transactionSet = true
			sawBusfileFlag = true
		case arg == "--transaction":
			if i+1 >= len(args) {
				return opts, true, fmt.Errorf("missing value for --transaction")
			}
			opts.transaction = args[i+1]
			opts.transactionSet = true
			sawBusfileFlag = true
			i++
		case strings.HasPrefix(arg, "--scope="):
			opts.scope = strings.TrimPrefix(arg, "--scope=")
			opts.scopeSet = true
			sawBusfileFlag = true
		case arg == "--scope":
			if i+1 >= len(args) {
				return opts, true, fmt.Errorf("missing value for --scope")
			}
			opts.scope = args[i+1]
			opts.scopeSet = true
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
	registerTestBusfileRunners(env)
	cfg := loadBusfileConfig()
	return runBusfilesWithExecutor(opts, env, stdin, stdout, stderr, cfg)
}

func runBusfilesWithExecutor(opts busfileOptions, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, cfg busfileConfig) int {
	if opts.transactionSet {
		cfg.transactionProvider = opts.transaction
	}
	if opts.scopeSet {
		cfg.transactionScope = opts.scope
	}

	if !isValidScope(cfg.transactionScope) {
		fmt.Fprintf(stderr, "bus: invalid usage: transaction scope %q is not supported\n", cfg.transactionScope)
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
	if err := validateBusfileCommands(commands); err != nil {
		var bfErr busfileError
		if errors.As(err, &bfErr) {
			fmt.Fprintln(stderr, bfErr.Error())
			return bfErr.ExitCode
		}
		fmt.Fprintf(stderr, "bus: %v\n", err)
		return 1
	}
	shellDispatchPaths, err := preflightDispatchTargets(commands, env, cfg)
	if err != nil {
		var bfErr busfileError
		if errors.As(err, &bfErr) {
			fmt.Fprintln(stderr, bfErr.Error())
			return bfErr.ExitCode
		}
		fmt.Fprintf(stderr, "bus: %v\n", err)
		return 1
	}
	resolvedProvider, warning, resolveErr := resolveTransactionProvider(cfg, commands, opts.transactionSet)
	if resolveErr != nil {
		fmt.Fprintln(stderr, resolveErr.Error())
		return 2
	}
	if warning != "" {
		fmt.Fprintln(stderr, warning)
	}
	cfg.transactionProvider = resolvedProvider

	if opts.check {
		return 0
	}

	switch cfg.transactionProvider {
	case "none":
		executor := hybridBusfileExecutor{
			shellLookupEnabled: cfg.shellLookupEnabled,
			resolvedTargets:    shellDispatchPaths,
		}
		return executeBusfileCommands(commands, opts, env, stdin, stdout, stderr, executor)
	case "fs":
		return executeBusfileCommandsFS(commands, opts, env, stdin, stdout, stderr, cfg)
	default:
		fmt.Fprintf(stderr, "bus: invalid usage: transaction provider %q is not implemented\n", cfg.transactionProvider)
		return 2
	}
}

func resolveTransactionProvider(cfg busfileConfig, commands []busfileCommand, cliTransactionSet bool) (provider string, warning string, err error) {
	if !isValidTransaction(cfg.transactionProvider) {
		return "", "", fmt.Errorf("bus: invalid usage: transaction provider %q is not supported", cfg.transactionProvider)
	}
	switch cfg.transactionProvider {
	case "none":
		return "none", "", nil
	case "fs":
		// fs provider requires in-process module dispatch to intercept writes.
		if allCommandsTxInProcess(commands) {
			return "fs", "", nil
		}
		if cliTransactionSet || !cfg.fallbackToNone {
			return "", "", fmt.Errorf("bus: transaction provider \"fs\" requires in-process tx runners for all targets")
		}
		return "none", "bus: warning: transaction provider \"fs\" requires in-process tx runners; falling back to \"none\"", nil
	default:
		if cliTransactionSet || !cfg.fallbackToNone {
			return "", "", fmt.Errorf("bus: invalid usage: transaction provider %q is not implemented", cfg.transactionProvider)
		}
		return "none", fmt.Sprintf("bus: warning: transaction provider %q unavailable; falling back to \"none\"", cfg.transactionProvider), nil
	}
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

func validateBusfileCommands(commands []busfileCommand) error {
	for _, command := range commands {
		if len(command.Argv) == 0 {
			return busfileError{
				File:     command.File,
				Line:     command.Line,
				Message:  "validation error: empty command",
				ExitCode: 1,
			}
		}
		if err := validateCommand(command.Argv); err != nil {
			return busfileError{
				File:     command.File,
				Line:     command.Line,
				Message:  "validation error: " + err.Error(),
				ExitCode: 1,
			}
		}
	}
	return nil
}

func preflightDispatchTargets(commands []busfileCommand, env []string, cfg busfileConfig) (map[string]string, error) {
	resolved := make(map[string]string)
	targetsToResolve := make(map[string]struct{})
	for _, command := range commands {
		if len(command.Argv) == 0 {
			continue
		}
		target := command.Argv[0]
		if hasInProcessRunner(target) || hasInProcessTxRunner(target) {
			continue
		}
		if !cfg.shellLookupEnabled {
			return nil, busfileError{
				File:     command.File,
				Line:     command.Line,
				Message:  fmt.Sprintf("dispatch error: shell lookup disabled and no in-process runner for target %q", target),
				ExitCode: 127,
			}
		}
		targetsToResolve[target] = struct{}{}
	}

	index := buildExecutableIndexForTargets(env, targetsToResolve)
	for _, command := range commands {
		if len(command.Argv) == 0 {
			continue
		}
		target := command.Argv[0]
		if hasInProcessRunner(target) || hasInProcessTxRunner(target) {
			continue
		}
		if path, ok := index[target]; ok {
			resolved[target] = path
			continue
		}
		return nil, busfileError{
			File:     command.File,
			Line:     command.Line,
			Message:  fmt.Sprintf("dispatch error: unknown target %q", target),
			ExitCode: 127,
		}
	}
	return resolved, nil
}

func hasInProcessRunner(target string) bool {
	_, ok := inProcessModuleRunners[target]
	return ok
}

func allCommandsInProcess(commands []busfileCommand) bool {
	if len(commands) == 0 {
		return false
	}
	for _, command := range commands {
		if len(command.Argv) == 0 {
			continue
		}
		if !hasInProcessRunner(command.Argv[0]) {
			return false
		}
	}
	return true
}

func hasInProcessTxRunner(target string) bool {
	_, ok := inProcessTxModuleRunners[target]
	return ok
}

func allCommandsTxInProcess(commands []busfileCommand) bool {
	if len(commands) == 0 {
		return false
	}
	for _, command := range commands {
		if len(command.Argv) == 0 {
			continue
		}
		if !hasInProcessTxRunner(command.Argv[0]) {
			return false
		}
	}
	return true
}

func executeBusfileCommands(commands []busfileCommand, opts busfileOptions, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, executor busfileExecutor) int {
	baseEnv := withBusBatchEnv(env)
	for _, command := range commands {
		if opts.trace {
			fmt.Fprintf(stdout, "%s:%d: bus %s\n", command.File, command.Line, strings.Join(command.Argv, " "))
		}
		commandEnv := withBusfileEnv(baseEnv, command.File, command.Line, "")
		code, runErr := executor.Execute(command, commandEnv, stdin, stdout, stderr)
		if runErr != nil {
			fmt.Fprintf(stderr, "%s:%d: %v\n", command.File, command.Line, runErr)
			return code
		}
	}
	return 0
}

func executeBusfileCommandsFS(commands []busfileCommand, opts busfileOptions, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, cfg busfileConfig) int {
	workspaceRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "bus: transaction provider \"fs\" failed to resolve workspace: %v\n", err)
		return 1
	}
	if err := recoverPendingFSTransactions(workspaceRoot, stderr); err != nil {
		fmt.Fprintf(stderr, "bus: transaction provider \"fs\" recovery failed: %v\n", err)
		return 1
	}

	var units [][]busfileCommand
	switch cfg.transactionScope {
	case "batch":
		units = [][]busfileCommand{commands}
	case "file":
		units = partitionCommandsByFile(commands)
	default:
		fmt.Fprintf(stderr, "bus: invalid usage: transaction scope %q is not supported\n", cfg.transactionScope)
		return 2
	}

	for _, unit := range units {
		if len(unit) == 0 {
			continue
		}
		code, runErr := executeFSUnit(unit, opts, env, stdin, stdout, stderr, workspaceRoot, cfg.transactionScope)
		if runErr != nil {
			if bf, ok := runErr.(busfileError); ok {
				fmt.Fprintln(stderr, bf.Error())
				return bf.ExitCode
			}
			fmt.Fprintln(stderr, runErr.Error())
			return code
		}
	}
	return 0
}

func partitionCommandsByFile(commands []busfileCommand) [][]busfileCommand {
	orderedFiles := make([]string, 0)
	grouped := map[string][]busfileCommand{}
	for _, command := range commands {
		if _, ok := grouped[command.File]; !ok {
			orderedFiles = append(orderedFiles, command.File)
		}
		grouped[command.File] = append(grouped[command.File], command)
	}
	units := make([][]busfileCommand, 0, len(orderedFiles))
	for _, file := range orderedFiles {
		units = append(units, grouped[file])
	}
	return units
}

func executeFSUnit(commands []busfileCommand, opts busfileOptions, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, workspaceRoot string, scope string) (int, error) {
	txID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	txRoot := filepath.Join(workspaceRoot, ".bus", "tx", txID)
	overlayRoot := filepath.Join(txRoot, "overlay")
	journalPath := filepath.Join(txRoot, "journal.json")

	if err := os.MkdirAll(filepath.Dir(overlayRoot), 0o755); err != nil {
		return 1, fmt.Errorf("bus: transaction provider \"fs\" failed to initialize: %v", err)
	}
	fsOverlay, err := txfs.New(workspaceRoot, overlayRoot)
	if err != nil {
		return 1, fmt.Errorf("bus: transaction provider \"fs\" failed to initialize: %v", err)
	}
	unitFiles := uniqueCommandFiles(commands)
	if scope == "batch" {
		if err := writeFSTxJournal(journalPath, fsTxJournal{
			State:   "begun",
			Scope:   scope,
			TxID:    txID,
			Files:   unitFiles,
			Updated: time.Now().UTC().Format(time.RFC3339Nano),
		}); err != nil {
			return 1, fmt.Errorf("bus: transaction provider \"fs\" failed to write journal: %v", err)
		}
	}

	baseEnv := withBusBatchEnv(env)
	for _, command := range commands {
		if opts.trace {
			fmt.Fprintf(stdout, "%s:%d: bus %s\n", command.File, command.Line, strings.Join(command.Argv, " "))
		}
		commandEnv := withBusfileEnv(baseEnv, command.File, command.Line, "fs")
		code, runErr := runBusfileCommandFS(command, commandEnv, stdin, stdout, stderr, fsOverlay)
		if runErr != nil {
			_ = fsOverlay.Rollback()
			_ = os.RemoveAll(txRoot)
			return code, busfileError{
				File:     command.File,
				Line:     command.Line,
				Message:  runErr.Error(),
				ExitCode: code,
			}
		}
	}

	if scope == "batch" {
		if err := writeFSTxJournal(journalPath, fsTxJournal{
			State:   "committing",
			Scope:   scope,
			TxID:    txID,
			Files:   unitFiles,
			Updated: time.Now().UTC().Format(time.RFC3339Nano),
		}); err != nil {
			_ = fsOverlay.Rollback()
			_ = os.RemoveAll(txRoot)
			return 1, fmt.Errorf("bus: transaction provider \"fs\" failed to write journal: %v", err)
		}
	}
	if err := fsOverlay.Commit(); err != nil {
		_ = fsOverlay.Rollback()
		_ = os.RemoveAll(txRoot)
		return 1, fmt.Errorf("bus: transaction provider \"fs\" commit failed: %v", err)
	}
	if scope == "batch" {
		if err := writeFSTxJournal(journalPath, fsTxJournal{
			State:   "committed",
			Scope:   scope,
			TxID:    txID,
			Files:   unitFiles,
			Updated: time.Now().UTC().Format(time.RFC3339Nano),
		}); err != nil {
			return 1, fmt.Errorf("bus: transaction provider \"fs\" failed to finalize journal: %v", err)
		}
	}
	if err := os.RemoveAll(txRoot); err != nil {
		return 1, fmt.Errorf("bus: transaction provider \"fs\" cleanup failed: %v", err)
	}
	return 0, nil
}

func writeFSTxJournal(path string, journal fsTxJournal) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func recoverPendingFSTransactions(workspaceRoot string, stderr io.Writer) error {
	txRoot := filepath.Join(workspaceRoot, ".bus", "tx")
	entries, err := os.ReadDir(txRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(txRoot, entry.Name())
		if !entry.IsDir() {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			fmt.Fprintf(stderr, "bus: warning: cleaned incomplete fs transaction artifact %s\n", path)
			continue
		}
		journalPath := filepath.Join(path, "journal.json")
		if _, err := os.Stat(journalPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		fmt.Fprintf(stderr, "bus: warning: recovered incomplete fs transaction %s by cleanup\n", entry.Name())
	}
	return nil
}

func uniqueCommandFiles(commands []busfileCommand) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(commands))
	for _, command := range commands {
		if _, ok := seen[command.File]; ok {
			continue
		}
		seen[command.File] = struct{}{}
		out = append(out, command.File)
	}
	sort.Strings(out)
	return out
}

func runBusfileCommandFS(command busfileCommand, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, fsOverlay *txfs.FS) (int, error) {
	if len(command.Argv) == 0 {
		return 65, fmt.Errorf("syntax error: empty command")
	}
	target := command.Argv[0]
	runner, ok := inProcessTxModuleRunners[target]
	if !ok {
		return 2, fmt.Errorf("dispatch error: transaction provider \"fs\" requires in-process tx runner for target %q", target)
	}
	code, err := runner(command.Argv[1:], env, stdin, stdout, stderr, fsOverlay)
	if err != nil {
		return 1, fmt.Errorf("dispatch error: %v", err)
	}
	if code != 0 {
		return code, fmt.Errorf("command failed (exit %d): %s", code, command.Raw)
	}
	return 0, nil
}

func runModuleViaTempWorkspaceAndMerge(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, fsOverlay *txfs.FS, runner inProcessModuleRunner) (int, error) {
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return 1, err
	}
	tmpRoot, err := os.MkdirTemp("", "bus-fs-module-*")
	if err != nil {
		return 1, err
	}
	defer os.RemoveAll(tmpRoot)
	snapshot, err := copyWorkspaceTree(workspaceRoot, tmpRoot)
	if err != nil {
		return 1, err
	}
	oldWD, err := os.Getwd()
	if err != nil {
		return 1, err
	}
	if err := os.Chdir(tmpRoot); err != nil {
		return 1, err
	}
	code, runErr := runner(args, env, stdin, stdout, stderr)
	_ = os.Chdir(oldWD)
	if runErr != nil {
		return 1, runErr
	}
	if code != 0 {
		return code, nil
	}
	if err := mergeWorkspaceChangesToTxFS(tmpRoot, fsOverlay, snapshot); err != nil {
		return 1, err
	}
	return 0, nil
}

func copyWorkspaceTree(srcRoot, dstRoot string) (map[string]workspaceFileInfo, error) {
	snapshot := make(map[string]workspaceFileInfo)
	err := filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if shouldIgnoreWorkspacePath(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		dstPath := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		snapshot[rel] = workspaceFileInfo{
			path:    dstPath,
			size:    info.Size(),
			modTime: info.ModTime().UnixNano(),
			mode:    info.Mode().Perm(),
		}
		return copyFile(srcRoot, rel, dstRoot, info.Mode().Perm())
	})
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

type workspaceFileInfo struct {
	path    string
	size    int64
	modTime int64
	mode    os.FileMode
}

func mergeWorkspaceChangesToTxFS(newRoot string, fsOverlay *txfs.FS, snapshot map[string]workspaceFileInfo) error {
	newFiles, err := listWorkspaceFiles(newRoot)
	if err != nil {
		return err
	}

	for rel, info := range newFiles {
		baseline, hadBase := snapshot[rel]
		if hadBase && baseline.size == info.size && baseline.modTime == info.modTime && baseline.mode == info.mode {
			continue
		}
		if err := writeIntoTxFS(fsOverlay, rel, info.path); err != nil {
			return err
		}
	}
	for rel := range snapshot {
		if _, ok := newFiles[rel]; ok {
			continue
		}
		if err := fsOverlay.Remove(rel); err != nil {
			return err
		}
	}
	return nil
}

func listWorkspaceFiles(root string) (map[string]workspaceFileInfo, error) {
	out := map[string]workspaceFileInfo{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if shouldIgnoreWorkspacePath(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		out[rel] = workspaceFileInfo{
			path:    path,
			size:    info.Size(),
			modTime: info.ModTime().UnixNano(),
			mode:    info.Mode().Perm(),
		}
		return nil
	})
	return out, err
}

func shouldIgnoreWorkspacePath(rel string) bool {
	if rel == ".git" || strings.HasPrefix(rel, ".git"+string(os.PathSeparator)) {
		return true
	}
	if rel == ".bus/tx" || strings.HasPrefix(rel, ".bus/tx"+string(os.PathSeparator)) {
		return true
	}
	return false
}

func copyFile(srcRoot, rel, dstRoot string, perm os.FileMode) error {
	src := filepath.Join(srcRoot, rel)
	dst := filepath.Join(dstRoot, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func writeIntoTxFS(fsOverlay *txfs.FS, relPath string, sourcePath string) error {
	in, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := fsOverlay.OpenFile(relPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func registerTestBusfileRunners(env []string) {
	value, ok := lookupEnvSlice(env, "BUS_TEST_ENABLE_TXWRITE")
	if !ok || value != "1" {
		return
	}
	if _, exists := inProcessModuleRunners["txwrite"]; !exists {
		inProcessModuleRunners["txwrite"] = func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
			if len(args) != 2 {
				return 2, fmt.Errorf("txwrite expects <path> <value>")
			}
			if args[1] == "fail" {
				return 1, nil
			}
			if err := os.MkdirAll(filepath.Dir(args[0]), 0o755); err != nil {
				return 1, err
			}
			f, err := os.OpenFile(args[0], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return 1, err
			}
			defer f.Close()
			if _, err := io.WriteString(f, args[1]+"\n"); err != nil {
				return 1, err
			}
			return 0, nil
		}
	}
	if _, exists := inProcessTxModuleRunners["txwrite"]; !exists {
		inProcessTxModuleRunners["txwrite"] = func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, fs *txfs.FS) (int, error) {
			if len(args) != 2 {
				return 2, fmt.Errorf("txwrite expects <path> <value>")
			}
			if args[1] == "fail" {
				return 1, nil
			}
			f, err := fs.OpenFile(args[0], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return 1, err
			}
			defer f.Close()
			if _, err := io.WriteString(f, args[1]+"\n"); err != nil {
				return 1, err
			}
			return 0, nil
		}
	}
}

func lookupEnvSlice(env []string, key string) (string, bool) {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix), true
		}
	}
	return "", false
}

func validateCommand(argv []string) error {
	if len(argv) < 2 {
		return nil
	}
	if argv[0] == "journal" && argv[1] == "add" {
		return validateJournalAdd(argv[2:])
	}
	if len(argv) >= 3 && argv[0] == "bank" && argv[1] == "add" && argv[2] == "transactions" {
		return validateBankAddTransactions(argv[3:])
	}
	return nil
}

func validateJournalAdd(args []string) error {
	var debitTotal decimalTotal
	var creditTotal decimalTotal
	hasDebit := false
	hasCredit := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--date" {
			if i+1 >= len(args) {
				return fmt.Errorf("journal add missing value for --date")
			}
			if !isISODate(args[i+1]) {
				return fmt.Errorf("journal add invalid date %q", args[i+1])
			}
			i++
			continue
		}
		if arg == "--debit" || arg == "--credit" {
			if i+1 >= len(args) {
				return fmt.Errorf("journal add missing value for %s", arg)
			}
			posting := args[i+1]
			account, amountText, ok := strings.Cut(posting, "=")
			if !ok || strings.TrimSpace(account) == "" || strings.TrimSpace(amountText) == "" {
				return fmt.Errorf("journal add invalid posting %q", posting)
			}
			amount, ok := parseDecimalAmount(amountText)
			if !ok {
				return fmt.Errorf("journal add invalid amount %q", amountText)
			}
			if amount.value.Sign() <= 0 {
				return fmt.Errorf("journal add amount must be positive: %q", amountText)
			}
			if arg == "--debit" {
				hasDebit = true
				debitTotal.Add(amount)
			} else {
				hasCredit = true
				creditTotal.Add(amount)
			}
			i++
		}
	}
	if !hasDebit || !hasCredit {
		return fmt.Errorf("journal add requires both debit and credit postings")
	}
	if debitTotal.Cmp(creditTotal) != 0 {
		return fmt.Errorf("journal add unbalanced entry: debit=%s credit=%s", debitTotal.String(), creditTotal.String())
	}
	return nil
}

func validateBankAddTransactions(args []string) error {
	for i := 0; i < len(args); i++ {
		if args[i] != "--set" {
			continue
		}
		if i+1 >= len(args) {
			return fmt.Errorf("bank add transactions missing value for --set")
		}
		key, value, ok := strings.Cut(args[i+1], "=")
		if !ok || strings.TrimSpace(key) == "" {
			return fmt.Errorf("bank add transactions invalid --set %q", args[i+1])
		}
		switch key {
		case "booked_date", "value_date":
			if value != "" && !isISODate(value) {
				return fmt.Errorf("bank add transactions invalid %s %q", key, value)
			}
		case "amount":
			if _, ok := parseDecimalAmount(value); !ok {
				return fmt.Errorf("bank add transactions invalid amount %q", value)
			}
		case "currency":
			if !isCurrencyCode(value) {
				return fmt.Errorf("bank add transactions invalid currency %q", value)
			}
		}
		i++
	}
	return nil
}

type decimalAmount struct {
	value *big.Int
	scale int
}

type decimalTotal struct {
	value big.Int
	scale int
}

func (t *decimalTotal) Add(amount decimalAmount) {
	if amount.scale > t.scale {
		factor := pow10Big(amount.scale - t.scale)
		t.value.Mul(&t.value, factor)
		t.scale = amount.scale
	}
	if amount.scale < t.scale {
		scaled := new(big.Int).Mul(amount.value, pow10Big(t.scale-amount.scale))
		t.value.Add(&t.value, scaled)
		return
	}
	t.value.Add(&t.value, amount.value)
}

func (t *decimalTotal) Cmp(other decimalTotal) int {
	switch {
	case t.scale == other.scale:
		return t.value.Cmp(&other.value)
	case t.scale > other.scale:
		scaledOther := new(big.Int).Mul(&other.value, pow10Big(t.scale-other.scale))
		return t.value.Cmp(scaledOther)
	default:
		scaledThis := new(big.Int).Mul(&t.value, pow10Big(other.scale-t.scale))
		return scaledThis.Cmp(&other.value)
	}
}

func (t *decimalTotal) String() string {
	s := t.value.String()
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = strings.TrimPrefix(s, "-")
	}
	if t.scale == 0 {
		if negative {
			return "-" + s
		}
		return s
	}
	if len(s) <= t.scale {
		s = strings.Repeat("0", t.scale-len(s)+1) + s
	}
	intPart := s[:len(s)-t.scale]
	fracPart := strings.TrimRight(s[len(s)-t.scale:], "0")
	if fracPart == "" {
		if negative {
			return "-" + intPart
		}
		return intPart
	}
	if negative {
		return "-" + intPart + "." + fracPart
	}
	return intPart + "." + fracPart
}

func parseDecimalAmount(text string) (decimalAmount, bool) {
	if text == "" {
		return decimalAmount{}, false
	}
	start := 0
	switch text[0] {
	case '+':
		start = 1
	case '-':
		start = 1
	}
	if start >= len(text) {
		return decimalAmount{}, false
	}
	digits := make([]byte, 0, len(text)-start)
	seenDot := false
	scale := 0
	seenDigit := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		switch {
		case ch >= '0' && ch <= '9':
			digits = append(digits, ch)
			if seenDot {
				scale++
			}
			seenDigit = true
		case ch == '.':
			if seenDot {
				return decimalAmount{}, false
			}
			seenDot = true
		default:
			return decimalAmount{}, false
		}
	}
	if !seenDigit {
		return decimalAmount{}, false
	}
	value := new(big.Int)
	if _, ok := value.SetString(string(digits), 10); !ok {
		return decimalAmount{}, false
	}
	if text[0] == '-' {
		value.Neg(value)
	}
	return decimalAmount{value: value, scale: scale}, true
}

var (
	decimalPow10Mu    sync.Mutex
	decimalPow10Cache = []*big.Int{big.NewInt(1)}
)

func pow10Big(power int) *big.Int {
	if power <= 0 {
		return decimalPow10Cache[0]
	}
	decimalPow10Mu.Lock()
	defer decimalPow10Mu.Unlock()
	for len(decimalPow10Cache) <= power {
		next := new(big.Int).Mul(decimalPow10Cache[len(decimalPow10Cache)-1], big.NewInt(10))
		decimalPow10Cache = append(decimalPow10Cache, next)
	}
	return decimalPow10Cache[power]
}

func isISODate(value string) bool {
	if value == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func isCurrencyCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, r := range value {
		if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
			return false
		}
	}
	return true
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
	logicalLine := make([]byte, 0, 256)
	logicalStartLine := 0
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Bytes()
		if len(logicalLine) == 0 {
			trimmed := strings.TrimSpace(string(line))
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			logicalStartLine = lineNo
			logicalLine = append(logicalLine, line...)
		} else {
			logicalLine = append(logicalLine, line...)
		}
		if hasLineContinuationBytes(logicalLine) {
			logicalLine = logicalLine[:len(logicalLine)-1]
			continue
		}

		trimmed := strings.TrimSpace(string(logicalLine))
		logicalLine = logicalLine[:0]
		logicalStart := logicalStartLine
		logicalStartLine = 0

		argv, parseErr := tokenizeBusLine(trimmed)
		if parseErr != nil {
			return busfileError{
				File:     path,
				Line:     logicalStart,
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
				Line:     logicalStart,
				Message:  "syntax error: empty command",
				ExitCode: 65,
			}
		}
		*commands = append(*commands, busfileCommand{
			File: path,
			Line: logicalStart,
			Raw:  trimmed,
			Argv: argv,
		})
	}
	if err := scanner.Err(); err != nil {
		return busfileError{File: path, Message: err.Error(), ExitCode: 2}
	}
	if len(logicalLine) > 0 {
		return busfileError{
			File:     path,
			Line:     logicalStartLine,
			Message:  "syntax error: line continuation at end of file",
			ExitCode: 65,
		}
	}
	return nil
}

func hasLineContinuation(line string) bool {
	return strings.HasSuffix(line, "\\")
}

func hasLineContinuationBytes(line []byte) bool {
	if len(line) == 0 {
		return false
	}
	return line[len(line)-1] == '\\'
}

func tokenizeBusLine(line string) ([]string, error) {
	tokens := make([]string, 0, 8)
	current := make([]byte, 0, len(line))
	inSingle := false
	inDouble := false
	escaped := false
	tokenStarted := false

	flush := func() {
		tokens = append(tokens, string(current))
		current = current[:0]
		tokenStarted = false
	}

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			current = append(current, ch)
			escaped = false
			tokenStarted = true
			continue
		}
		if inSingle {
			if ch == '\'' {
				inSingle = false
				continue
			}
			current = append(current, ch)
			tokenStarted = true
			continue
		}
		if inDouble {
			if ch == '"' {
				inDouble = false
				continue
			}
			current = append(current, ch)
			tokenStarted = true
			continue
		}

		switch ch {
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
			return nil, fmt.Errorf("disallowed token %q", string(ch))
		default:
			current = append(current, ch)
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

func runBusfileCommand(command busfileCommand, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, resolvedTargets map[string]string) (int, error) {
	if len(command.Argv) == 0 {
		return 65, fmt.Errorf("syntax error: empty command")
	}
	subcommand := command.Argv[0]
	executable := "bus-" + subcommand

	path := ""
	if resolvedTargets != nil {
		path = resolvedTargets[subcommand]
	}
	if path == "" {
		var err error
		path, err = lookPathEnv(executable, env)
		if err != nil {
			return 127, fmt.Errorf("dispatch error: unknown target %q", subcommand)
		}
		if resolvedTargets != nil {
			resolvedTargets[subcommand] = path
		}
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

func withBusBatchEnv(env []string) []string {
	return withBusfileEnv(env, "", 0, "")
}

func withBusfileEnv(env []string, file string, line int, provider string) []string {
	updated := make([]string, 0, len(env)+4)
	foundBatch := false
	foundFile := false
	foundLine := false
	foundProvider := provider == ""
	lineValue := strconv.Itoa(line)

	for _, entry := range env {
		switch {
		case strings.HasPrefix(entry, "BUS_BATCH="):
			updated = append(updated, "BUS_BATCH=1")
			foundBatch = true
		case strings.HasPrefix(entry, "BUS_BUSFILE="):
			if file != "" {
				updated = append(updated, "BUS_BUSFILE="+file)
				foundFile = true
			} else {
				updated = append(updated, entry)
			}
		case strings.HasPrefix(entry, "BUS_BUSFILE_LINE="):
			if line > 0 {
				updated = append(updated, "BUS_BUSFILE_LINE="+lineValue)
				foundLine = true
			} else {
				updated = append(updated, entry)
			}
		case strings.HasPrefix(entry, "BUS_TRANSACTION_PROVIDER="):
			if provider != "" {
				updated = append(updated, "BUS_TRANSACTION_PROVIDER="+provider)
				foundProvider = true
			} else {
				updated = append(updated, entry)
			}
		default:
			updated = append(updated, entry)
		}
	}

	if !foundBatch {
		updated = append(updated, "BUS_BATCH=1")
	}
	if file != "" && !foundFile {
		updated = append(updated, "BUS_BUSFILE="+file)
	}
	if line > 0 && !foundLine {
		updated = append(updated, "BUS_BUSFILE_LINE="+lineValue)
	}
	if provider != "" && !foundProvider {
		updated = append(updated, "BUS_TRANSACTION_PROVIDER="+provider)
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
	case "none", "fs", "git", "snapshot", "copy":
		return true
	default:
		return false
	}
}

func loadBusfileConfig() busfileConfig {
	cfg := busfileConfig{
		transactionProvider: "none",
		transactionScope:    "file",
		fallbackToNone:      true,
		validationLevel:     "syntax",
		shellLookupEnabled:  true,
	}
	applyDatapackageConfig(&cfg)
	applyPreferencesConfig(&cfg)
	return cfg
}

func applyDatapackageConfig(cfg *busfileConfig) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	path := filepath.Join(wd, "datapackage.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return
	}
	busMap, ok := mapValue(doc["bus"])
	if !ok {
		return
	}
	busfileMap, ok := mapValue(busMap["busfile"])
	if !ok {
		return
	}
	if txMap, ok := mapValue(busfileMap["transaction"]); ok {
		if provider, ok := stringValue(txMap["provider"]); ok && provider != "" {
			cfg.transactionProvider = provider
		}
		if scope, ok := stringValue(txMap["scope"]); ok && scope != "" {
			cfg.transactionScope = scope
		}
		if fallback, ok := boolValue(txMap["fallback_to_none"]); ok {
			cfg.fallbackToNone = fallback
		}
	}
	if dispatchMap, ok := mapValue(busfileMap["dispatch"]); ok {
		if enabled, ok := boolValue(dispatchMap["shell_lookup_enabled"]); ok {
			cfg.shellLookupEnabled = enabled
		}
	}
	if validationMap, ok := mapValue(busfileMap["validation"]); ok {
		if level, ok := stringValue(validationMap["level"]); ok && level != "" {
			cfg.validationLevel = level
		}
	}
}

func applyPreferencesConfig(cfg *busfileConfig) {
	path := preferencesPath()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var env struct {
		Values map[string]json.RawMessage `json:"values"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return
	}
	if len(env.Values) == 0 {
		return
	}
	applyPreferencesObject(env.Values["bus.busfile"], cfg)
	readPrefString(env.Values, "bus.busfile.transaction.provider", &cfg.transactionProvider)
	readPrefString(env.Values, "bus.busfile.transaction.scope", &cfg.transactionScope)
	readPrefBool(env.Values, "bus.busfile.transaction.fallback_to_none", &cfg.fallbackToNone)
	readPrefString(env.Values, "bus.busfile.validation.level", &cfg.validationLevel)
	readPrefBool(env.Values, "bus.busfile.dispatch.shell_lookup_enabled", &cfg.shellLookupEnabled)
}

func applyPreferencesObject(raw json.RawMessage, cfg *busfileConfig) {
	if len(raw) == 0 {
		return
	}
	var busfileMap map[string]any
	if err := json.Unmarshal(raw, &busfileMap); err != nil {
		return
	}
	if txMap, ok := mapValue(busfileMap["transaction"]); ok {
		if provider, ok := stringValue(txMap["provider"]); ok && provider != "" {
			cfg.transactionProvider = provider
		}
		if scope, ok := stringValue(txMap["scope"]); ok && scope != "" {
			cfg.transactionScope = scope
		}
		if fallback, ok := boolValue(txMap["fallback_to_none"]); ok {
			cfg.fallbackToNone = fallback
		}
	}
	if dispatchMap, ok := mapValue(busfileMap["dispatch"]); ok {
		if enabled, ok := boolValue(dispatchMap["shell_lookup_enabled"]); ok {
			cfg.shellLookupEnabled = enabled
		}
	}
	if validationMap, ok := mapValue(busfileMap["validation"]); ok {
		if level, ok := stringValue(validationMap["level"]); ok && level != "" {
			cfg.validationLevel = level
		}
	}
}

func readPrefString(values map[string]json.RawMessage, key string, dest *string) {
	raw, ok := values[key]
	if !ok || len(raw) == 0 {
		return
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return
	}
	if value != "" {
		*dest = value
	}
}

func readPrefBool(values map[string]json.RawMessage, key string, dest *bool) {
	raw, ok := values[key]
	if !ok || len(raw) == 0 {
		return
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return
	}
	*dest = value
}

func preferencesPath() string {
	if p := os.Getenv("BUS_PREFERENCES_PATH"); p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		dir := os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		if dir == "" {
			return ""
		}
		return filepath.Join(dir, "BusDK", "preferences.json")
	}
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "busdk", "preferences.json")
}

func mapValue(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

func stringValue(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

func boolValue(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
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
	fmt.Fprintln(stderr, "tip: did you mean `bus shell`?")
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
	fmt.Fprintln(stdout, "tip: use `bus shell` for interactive command entry")
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

	exts := windowsExtSet(env)

	seen := map[string]struct{}{}
	for _, dir := range filepath.SplitList(pathValue) {
		if dir == "" {
			dir = "."
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasPrefix(name, "bus-") {
				continue
			}
			if !entryExecutable(entry) {
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
	}

	if _, hasValidate := seen["validate"]; hasValidate {
		if _, hasAudit := seen["audit"]; !hasAudit {
			seen["audit"] = struct{}{}
		}
	}
	subcommands := make([]string, 0, len(seen))
	for name := range seen {
		subcommands = append(subcommands, name)
	}
	sort.Strings(subcommands)
	return subcommands
}

func windowsExtSet(env []string) map[string]struct{} {
	if runtime.GOOS != "windows" {
		return nil
	}
	exts := map[string]struct{}{}
	for _, ext := range windowsPathExts(env) {
		if ext == "" {
			continue
		}
		exts[strings.ToLower(ext)] = struct{}{}
	}
	return exts
}

func buildExecutableIndexForTargets(env []string, targets map[string]struct{}) map[string]string {
	result := make(map[string]string, len(targets))
	if len(targets) == 0 {
		return result
	}
	pathValue, _ := lookupEnv(env, "PATH")
	if pathValue == "" {
		return result
	}

	needed := make(map[string]struct{}, len(targets))
	for target := range targets {
		needed["bus-"+target] = struct{}{}
	}
	exts := windowsExtSet(env)

	for _, dir := range filepath.SplitList(pathValue) {
		if len(needed) == 0 {
			break
		}
		if dir == "" {
			dir = "."
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if len(needed) == 0 {
				break
			}
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			commandName := name
			if runtime.GOOS == "windows" {
				ext := strings.ToLower(filepath.Ext(name))
				if ext == "" {
					continue
				}
				if _, ok := exts[ext]; !ok {
					continue
				}
				commandName = strings.TrimSuffix(name, ext)
			}
			if _, ok := needed[commandName]; !ok {
				continue
			}
			if !entryExecutable(entry) {
				continue
			}
			fullPath := filepath.Join(dir, name)
			result[strings.TrimPrefix(commandName, "bus-")] = fullPath
			delete(needed, commandName)
		}
	}
	return result
}

func entryExecutable(entry os.DirEntry) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	info, err := entry.Info()
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	return info.Mode().Perm()&0o111 != 0
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
