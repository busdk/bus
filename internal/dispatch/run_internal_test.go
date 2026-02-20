package dispatch

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bus/internal/txfs"
)

func TestRunBusfileShellLookupDisabledUsesInProcessRunner(t *testing.T) {
	tempDir := t.TempDir()
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

	restore := setInProcessRunnerForTest("accounts", func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
		_, _ = io.WriteString(stdout, "INPROC:"+strings.Join(args, " ")+"\n")
		return 0, nil
	})
	defer restore()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	withChdirInternal(t, tempDir, func() {
		code := Run([]string{"bus", busfile}, os.Environ(), nil, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
		}
	})
	if stdout.String() != "INPROC:list\n" {
		t.Fatalf("expected in-process output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunBusfileFSProviderUsesInProcessRunners(t *testing.T) {
	tempDir := t.TempDir()
	busfile := filepath.Join(tempDir, "2024-02.bus")
	if err := os.WriteFile(busfile, []byte("accounts list\n"), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"file","fallback_to_none":false}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	restore := setInProcessRunnerForTest("accounts", func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
		_, _ = io.WriteString(stdout, "INPROC:"+strings.Join(args, " ")+"\n")
		return 0, nil
	})
	defer restore()
	restoreTx := setInProcessTxRunnerForTest("accounts", func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, fs *txfs.FS) (int, error) {
		_, _ = io.WriteString(stdout, "INPROC:"+strings.Join(args, " ")+"\n")
		return 0, nil
	})
	defer restoreTx()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	withChdirInternal(t, tempDir, func() {
		code := Run([]string{"bus", busfile}, os.Environ(), nil, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
		}
	})
	if stdout.String() != "INPROC:list\n" {
		t.Fatalf("expected in-process output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func setInProcessRunnerForTest(target string, runner inProcessModuleRunner) func() {
	prev, had := inProcessModuleRunners[target]
	inProcessModuleRunners[target] = runner
	return func() {
		if had {
			inProcessModuleRunners[target] = prev
			return
		}
		delete(inProcessModuleRunners, target)
	}
}

func setInProcessTxRunnerForTest(target string, runner inProcessTxModuleRunner) func() {
	prev, had := inProcessTxModuleRunners[target]
	inProcessTxModuleRunners[target] = runner
	return func() {
		if had {
			inProcessTxModuleRunners[target] = prev
			return
		}
		delete(inProcessTxModuleRunners, target)
	}
}

func TestRunBusfileFSProviderRollbackOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	busfile := filepath.Join(tempDir, "2024-02.bus")
	content := "txwrite file.txt first\n" +
		"txwrite file.txt fail\n"
	if err := os.WriteFile(busfile, []byte(content), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"batch","fallback_to_none":false}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	restoreTx := setInProcessTxRunnerForTest("txwrite", testTxWriteRunner())
	defer restoreTx()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	withChdirInternal(t, tempDir, func() {
		code := Run([]string{"bus", busfile}, os.Environ(), nil, &stdout, &stderr)
		if code != 1 {
			t.Fatalf("expected exit 1, got %d (stderr: %q)", code, stderr.String())
		}
	})
	if _, err := os.Stat(filepath.Join(tempDir, "file.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected rollback to remove file, got stat err=%v", err)
	}
	if !strings.Contains(stderr.String(), "command failed (exit 1)") {
		t.Fatalf("expected command failure, got %q", stderr.String())
	}
}

func TestRunBusfileFSProviderBatchCommitOnSuccess(t *testing.T) {
	tempDir := t.TempDir()
	busfile := filepath.Join(tempDir, "2024-02.bus")
	content := "txwrite file.txt one\n" +
		"txwrite file.txt two\n"
	if err := os.WriteFile(busfile, []byte(content), 0o600); err != nil {
		t.Fatalf("write busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"batch","fallback_to_none":false}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	restoreTx := setInProcessTxRunnerForTest("txwrite", testTxWriteRunner())
	defer restoreTx()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	withChdirInternal(t, tempDir, func() {
		code := Run([]string{"bus", busfile}, os.Environ(), nil, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("expected exit 0, got %d (stderr: %q)", code, stderr.String())
		}
	})
	body, err := os.ReadFile(filepath.Join(tempDir, "file.txt"))
	if err != nil {
		t.Fatalf("read committed file: %v", err)
	}
	if string(body) != "one\ntwo\n" {
		t.Fatalf("unexpected committed content %q", string(body))
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunBusfileFSProviderFileScopeCommitsPerFile(t *testing.T) {
	tempDir := t.TempDir()
	first := filepath.Join(tempDir, "a.bus")
	second := filepath.Join(tempDir, "b.bus")
	if err := os.WriteFile(first, []byte("txwrite file.txt one\n"), 0o600); err != nil {
		t.Fatalf("write first busfile: %v", err)
	}
	if err := os.WriteFile(second, []byte("txwrite file.txt fail\n"), 0o600); err != nil {
		t.Fatalf("write second busfile: %v", err)
	}
	datapackage := `{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"file","fallback_to_none":false}}}}`
	if err := os.WriteFile(filepath.Join(tempDir, "datapackage.json"), []byte(datapackage), 0o600); err != nil {
		t.Fatalf("write datapackage: %v", err)
	}

	restoreTx := setInProcessTxRunnerForTest("txwrite", testTxWriteRunner())
	defer restoreTx()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	withChdirInternal(t, tempDir, func() {
		code := Run([]string{"bus", first, second}, os.Environ(), nil, &stdout, &stderr)
		if code != 1 {
			t.Fatalf("expected exit 1, got %d (stderr: %q)", code, stderr.String())
		}
	})
	body, err := os.ReadFile(filepath.Join(tempDir, "file.txt"))
	if err != nil {
		t.Fatalf("read file after partial commit: %v", err)
	}
	if string(body) != "one\n" {
		t.Fatalf("expected first file committed before second fails, got %q", string(body))
	}
}

func testTxWriteRunner() inProcessTxModuleRunner {
	return func(args []string, env []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, fs *txfs.FS) (int, error) {
		if len(args) != 2 {
			return 2, errors.New("expected path and value")
		}
		path := args[0]
		value := args[1]
		if value == "fail" {
			return 1, nil
		}
		f, err := fs.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return 1, err
		}
		defer f.Close()
		if _, err := io.WriteString(f, value+"\n"); err != nil {
			return 1, err
		}
		return 0, nil
	}
}

func withChdirInternal(t *testing.T, dir string, fn func()) {
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
