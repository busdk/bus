package dispatch

import (
	"bus/internal/txfs"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var benchEnvOut []string

type noOpBusfileExecutor struct{}

func (noOpBusfileExecutor) Execute(command busfileCommand, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	return 0, nil
}

func BenchmarkWithBusfileEnv(b *testing.B) {
	env := make([]string, 0, 96)
	for i := 0; i < cap(env); i++ {
		env = append(env, fmt.Sprintf("KEY_%03d=value_%03d", i, i))
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchEnvOut = withBusfileEnv(env, "bench.bus", i, "")
	}
}

func BenchmarkExecuteBusfileCommandsEnvOverhead(b *testing.B) {
	env := make([]string, 0, 96)
	for i := 0; i < cap(env); i++ {
		env = append(env, fmt.Sprintf("KEY_%03d=value_%03d", i, i))
	}
	commands := make([]busfileCommand, 128)
	for i := range commands {
		commands[i] = busfileCommand{
			File: "bench.bus",
			Line: i + 1,
			Raw:  "bank list",
			Argv: []string{"bank", "list"},
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if code := executeBusfileCommands(commands, busfileOptions{}, env, nil, io.Discard, io.Discard, noOpBusfileExecutor{}); code != 0 {
			b.Fatalf("unexpected exit code: %d", code)
		}
	}
}

func BenchmarkPreflightDispatchTargetsRepeatedLookups(b *testing.B) {
	tempDir := b.TempDir()
	exe := filepath.Join(tempDir, "bus-accounts")
	src := []byte("#!/bin/sh\nexit 0\n")
	if runtime.GOOS == "windows" {
		exe = filepath.Join(tempDir, "bus-accounts.exe")
		src = []byte("@echo off\r\nexit /b 0\r\n")
	}
	if err := os.WriteFile(exe, src, 0o755); err != nil {
		b.Fatalf("write fake command: %v", err)
	}

	env := []string{"PATH=" + tempDir}
	cfg := busfileConfig{shellLookupEnabled: true}
	commands := make([]busfileCommand, 256)
	for i := range commands {
		commands[i] = busfileCommand{
			File: "bench.bus",
			Line: i + 1,
			Raw:  "accounts list",
			Argv: []string{"accounts", "list"},
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := preflightDispatchTargets(commands, env, cfg); err != nil {
			b.Fatalf("preflight failed: %v", err)
		}
	}
}

func BenchmarkPreflightDispatchTargetsUniqueLookups(b *testing.B) {
	tempDir := b.TempDir()
	env := []string{"PATH=" + tempDir}
	cfg := busfileConfig{shellLookupEnabled: true}

	commands := make([]busfileCommand, 256)
	for i := range commands {
		target := fmt.Sprintf("bench-%03d", i)
		exe := filepath.Join(tempDir, "bus-"+target)
		src := []byte("#!/bin/sh\nexit 0\n")
		if runtime.GOOS == "windows" {
			exe += ".exe"
			src = []byte("@echo off\r\nexit /b 0\r\n")
		}
		if err := os.WriteFile(exe, src, 0o755); err != nil {
			b.Fatalf("write fake command: %v", err)
		}
		commands[i] = busfileCommand{
			File: "bench.bus",
			Line: i + 1,
			Raw:  target + " list",
			Argv: []string{target, "list"},
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := preflightDispatchTargets(commands, env, cfg); err != nil {
			b.Fatalf("preflight failed: %v", err)
		}
	}
}

func BenchmarkPreflightDispatchTargetsUniqueLookupsWidePath(b *testing.B) {
	const pathDirs = 64
	dirs := make([]string, 0, pathDirs)
	commands := make([]busfileCommand, 0, pathDirs)
	for i := 0; i < pathDirs; i++ {
		dir := filepath.Join(b.TempDir(), fmt.Sprintf("path-%03d", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("mkdir path dir: %v", err)
		}
		target := fmt.Sprintf("wide-%03d", i)
		exe := filepath.Join(dir, "bus-"+target)
		src := []byte("#!/bin/sh\nexit 0\n")
		if runtime.GOOS == "windows" {
			exe += ".exe"
			src = []byte("@echo off\r\nexit /b 0\r\n")
		}
		if err := os.WriteFile(exe, src, 0o755); err != nil {
			b.Fatalf("write fake command: %v", err)
		}
		dirs = append(dirs, dir)
		commands = append(commands, busfileCommand{
			File: "bench.bus",
			Line: i + 1,
			Raw:  target + " list",
			Argv: []string{target, "list"},
		})
	}

	env := []string{"PATH=" + strings.Join(dirs, string(os.PathListSeparator))}
	cfg := busfileConfig{shellLookupEnabled: true}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := preflightDispatchTargets(commands, env, cfg); err != nil {
			b.Fatalf("preflight failed: %v", err)
		}
	}
}

func BenchmarkValidateBusfileCommandsJournalAdd(b *testing.B) {
	commands := make([]busfileCommand, 256)
	for i := range commands {
		commands[i] = busfileCommand{
			File: "bench.bus",
			Line: i + 1,
			Raw:  "journal add --date 2026-02-22 --debit assets:cash=123.45 --credit income:sales=123.45",
			Argv: []string{
				"journal", "add",
				"--date", "2026-02-22",
				"--debit", "assets:cash=123.45",
				"--credit", "income:sales=123.45",
			},
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validateBusfileCommands(commands); err != nil {
			b.Fatalf("validate failed: %v", err)
		}
	}
}

func BenchmarkValidateBusfileCommandsBankAddTransactions(b *testing.B) {
	commands := make([]busfileCommand, 256)
	for i := range commands {
		commands[i] = busfileCommand{
			File: "bench.bus",
			Line: i + 1,
			Raw:  "bank add transactions --set amount=123.45 --set currency=EUR --set booked_date=2026-02-22",
			Argv: []string{
				"bank", "add", "transactions",
				"--set", "amount=123.45",
				"--set", "currency=EUR",
				"--set", "booked_date=2026-02-22",
			},
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validateBusfileCommands(commands); err != nil {
			b.Fatalf("validate failed: %v", err)
		}
	}
}

func BenchmarkValidateJournalAddSingle(b *testing.B) {
	args := []string{
		"--date", "2026-02-22",
		"--debit", "assets:cash=123.45",
		"--credit", "income:sales=123.45",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validateJournalAdd(args); err != nil {
			b.Fatalf("validate failed: %v", err)
		}
	}
}

func BenchmarkValidateBankAddTransactionsSingle(b *testing.B) {
	args := []string{
		"--set", "amount=123.45",
		"--set", "currency=EUR",
		"--set", "booked_date=2026-02-22",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validateBankAddTransactions(args); err != nil {
			b.Fatalf("validate failed: %v", err)
		}
	}
}

func BenchmarkValidateBankAddTransactionsDateFieldsOnlySingle(b *testing.B) {
	args := []string{
		"--set", "booked_date=2026-02-22",
		"--set", "value_date=2026-02-23",
		"--set", "currency=EUR",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validateBankAddTransactions(args); err != nil {
			b.Fatalf("validate failed: %v", err)
		}
	}
}

func BenchmarkValidateBusfileCommandsBankAddTransactionsDateFieldsOnly(b *testing.B) {
	commands := make([]busfileCommand, 256)
	for i := range commands {
		commands[i] = busfileCommand{
			File: "bench.bus",
			Line: i + 1,
			Raw:  "bank add transactions --set booked_date=2026-02-22 --set value_date=2026-02-23 --set currency=EUR",
			Argv: []string{
				"bank", "add", "transactions",
				"--set", "booked_date=2026-02-22",
				"--set", "value_date=2026-02-23",
				"--set", "currency=EUR",
			},
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := validateBusfileCommands(commands); err != nil {
			b.Fatalf("validate failed: %v", err)
		}
	}
}

func BenchmarkIsISODate(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !isISODate("2026-02-22") {
			b.Fatalf("expected valid date")
		}
	}
}

func BenchmarkTokenizeBusLineJournalAdd(b *testing.B) {
	line := `journal add --date 2026-02-22 --debit "assets:cash=123.45" --credit income:sales=123.45`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tokens, err := tokenizeBusLine(line)
		if err != nil {
			b.Fatalf("tokenize failed: %v", err)
		}
		if len(tokens) == 0 {
			b.Fatalf("expected tokens")
		}
	}
}

func BenchmarkCollectBusfileCommands(b *testing.B) {
	tempDir := b.TempDir()
	busfile := filepath.Join(tempDir, "bench.bus")

	var content strings.Builder
	for i := 0; i < 512; i++ {
		content.WriteString("journal add --date 2026-02-22 --debit assets:cash=123.45 --credit income:sales=123.45\n")
	}
	if err := os.WriteFile(busfile, []byte(content.String()), 0o600); err != nil {
		b.Fatalf("write busfile: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		commands := make([]busfileCommand, 0, 512)
		if err := collectBusfileCommands(busfile, map[string]bool{}, &commands); err != nil {
			b.Fatalf("collect failed: %v", err)
		}
		if len(commands) != 512 {
			b.Fatalf("expected 512 commands, got %d", len(commands))
		}
	}
}

func BenchmarkListSubcommandsDensePath(b *testing.B) {
	tempDir := b.TempDir()
	for i := 0; i < 256; i++ {
		name := filepath.Join(tempDir, fmt.Sprintf("bus-cmd-%03d", i))
		body := []byte("#!/bin/sh\nexit 0\n")
		if runtime.GOOS == "windows" {
			name += ".exe"
			body = []byte("@echo off\r\nexit /b 0\r\n")
		}
		if err := os.WriteFile(name, body, 0o755); err != nil {
			b.Fatalf("write fake subcommand: %v", err)
		}
	}
	for i := 0; i < 1024; i++ {
		name := filepath.Join(tempDir, fmt.Sprintf("noise-%04d", i))
		if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
			b.Fatalf("write noise file: %v", err)
		}
	}

	env := []string{"PATH=" + tempDir}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		got := listSubcommands(env)
		if len(got) != 256 {
			b.Fatalf("expected 256 subcommands, got %d", len(got))
		}
	}
}

func BenchmarkMergeWorkspaceChangesToTxFSUnchangedTree(b *testing.B) {
	for _, tc := range []struct {
		name  string
		files int
		size  int
	}{
		{name: "files_64_size_4096", files: 64, size: 4 * 1024},
		{name: "files_128_size_65536", files: 128, size: 64 * 1024},
	} {
		b.Run(tc.name, func(b *testing.B) {
			baseRoot := filepath.Join(b.TempDir(), "base")
			newRoot := filepath.Join(b.TempDir(), "new")
			overlayBase := filepath.Join(b.TempDir(), "overlay")

			payload := bytesRepeat('x', tc.size)
			for i := 0; i < tc.files; i++ {
				rel := filepath.Join("dir", fmt.Sprintf("%04d", i), "file.bin")
				basePath := filepath.Join(baseRoot, rel)
				newPath := filepath.Join(newRoot, rel)
				if err := os.MkdirAll(filepath.Dir(basePath), 0o755); err != nil {
					b.Fatalf("mkdir base: %v", err)
				}
				if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
					b.Fatalf("mkdir new: %v", err)
				}
				if err := os.WriteFile(basePath, payload, 0o644); err != nil {
					b.Fatalf("write base: %v", err)
				}
				if err := os.WriteFile(newPath, payload, 0o644); err != nil {
					b.Fatalf("write new: %v", err)
				}
			}

			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				overlayRoot := filepath.Join(overlayBase, fmt.Sprintf("run-%d", i))
				fsOverlay, err := txfs.New(baseRoot, overlayRoot)
				if err != nil {
					b.Fatalf("new txfs: %v", err)
				}
				b.StartTimer()

				snapshot, err := listWorkspaceFiles(newRoot)
				if err != nil {
					b.Fatalf("snapshot failed: %v", err)
				}
				if err := mergeWorkspaceChangesToTxFS(newRoot, fsOverlay, snapshot); err != nil {
					b.Fatalf("merge failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkRunModuleViaTempWorkspaceAndMerge(b *testing.B) {
	for _, tc := range []struct {
		name   string
		files  int
		size   int
		mutate bool
	}{
		{name: "unchanged_files_32_size_4096", files: 32, size: 4 * 1024, mutate: false},
		{name: "mutate_one_file_files_32_size_4096", files: 32, size: 4 * 1024, mutate: true},
		{name: "unchanged_files_64_size_16384", files: 64, size: 16 * 1024, mutate: false},
		{name: "mutate_one_file_files_64_size_16384", files: 64, size: 16 * 1024, mutate: true},
	} {
		b.Run(tc.name, func(b *testing.B) {
			workspaceRoot := b.TempDir()
			overlayBase := b.TempDir()
			payload := bytesRepeat('x', tc.size)
			mutated := bytesRepeat('y', tc.size)
			mutatedRel := filepath.Join("dir", "0000", "file.bin")

			for i := 0; i < tc.files; i++ {
				rel := filepath.Join("dir", fmt.Sprintf("%04d", i), "file.bin")
				path := filepath.Join(workspaceRoot, rel)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					b.Fatalf("mkdir workspace: %v", err)
				}
				if err := os.WriteFile(path, payload, 0o644); err != nil {
					b.Fatalf("write workspace file: %v", err)
				}
			}

			oldWD, err := os.Getwd()
			if err != nil {
				b.Fatalf("getwd: %v", err)
			}
			if err := os.Chdir(workspaceRoot); err != nil {
				b.Fatalf("chdir workspace: %v", err)
			}
			b.Cleanup(func() {
				_ = os.Chdir(oldWD)
			})

			runner := func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
				if !tc.mutate {
					return 0, nil
				}
				return 0, os.WriteFile(mutatedRel, mutated, 0o644)
			}

			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				overlayRoot := filepath.Join(overlayBase, fmt.Sprintf("run-%d", i))
				fsOverlay, err := txfs.New(workspaceRoot, overlayRoot)
				if err != nil {
					b.Fatalf("new txfs: %v", err)
				}
				code, err := runModuleViaTempWorkspaceAndMerge(nil, nil, nil, io.Discard, io.Discard, fsOverlay, runner)
				if err != nil {
					b.Fatalf("runModuleViaTempWorkspaceAndMerge failed: %v", err)
				}
				if code != 0 {
					b.Fatalf("unexpected exit code: %d", code)
				}
			}
		})
	}
}

func bytesRepeat(ch byte, n int) []byte {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = ch
	}
	return buf
}
