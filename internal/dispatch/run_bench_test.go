package dispatch

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
		benchEnvOut = withBusfileEnv(env, "bench.bus", i)
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
		if err := preflightDispatchTargets(commands, env, cfg); err != nil {
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
