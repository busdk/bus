package txfs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFallsBackToBase(t *testing.T) {
	root := t.TempDir()
	overlay := filepath.Join(t.TempDir(), "overlay")
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs, err := New(root, overlay)
	if err != nil {
		t.Fatal(err)
	}
	f, err := fs.Open("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); !strings.Contains(got, "base") {
		t.Fatalf("expected base content, got %q", got)
	}
}

func TestAppendCopiesBaseThenWritesOverlay(t *testing.T) {
	root := t.TempDir()
	overlay := filepath.Join(t.TempDir(), "overlay")
	basePath := filepath.Join(root, "a.txt")
	if err := os.WriteFile(basePath, []byte("line1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs, err := New(root, overlay)
	if err != nil {
		t.Fatal(err)
	}
	f, err := fs.OpenFile("a.txt", os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("line2\n"); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	ov, err := os.ReadFile(filepath.Join(overlay, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(ov) != "line1\nline2\n" {
		t.Fatalf("unexpected overlay content: %q", string(ov))
	}
	base, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(base) != "line1\n" {
		t.Fatalf("expected base unchanged before commit, got %q", string(base))
	}
	if err := fs.Commit(); err != nil {
		t.Fatal(err)
	}
	base, err = os.ReadFile(basePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(base) != "line1\nline2\n" {
		t.Fatalf("expected committed content, got %q", string(base))
	}
}

func TestRemoveAndCommitDeletesBase(t *testing.T) {
	root := t.TempDir()
	overlay := filepath.Join(t.TempDir(), "overlay")
	path := filepath.Join(root, "delete.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs, err := New(root, overlay)
	if err != nil {
		t.Fatal(err)
	}
	if err := fs.Remove("delete.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Open("delete.txt"); !os.IsNotExist(err) {
		t.Fatalf("expected not exist before commit, got %v", err)
	}
	if err := fs.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected deleted after commit, got %v", err)
	}
}

func TestCommitLargeFileStreamingPath(t *testing.T) {
	root := t.TempDir()
	overlay := filepath.Join(t.TempDir(), "overlay")
	basePath := filepath.Join(root, "big.csv")
	f, err := os.Create(basePath)
	if err != nil {
		t.Fatal(err)
	}
	w := bufio.NewWriter(f)
	for i := 0; i < 100000; i++ {
		if _, err := fmt.Fprintf(w, "%d,alpha,beta\n", i); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	fs, err := New(root, overlay)
	if err != nil {
		t.Fatal(err)
	}
	out, err := fs.OpenFile("big.csv", os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := out.WriteString("100000,omega,zeta\n"); err != nil {
		t.Fatal(err)
	}
	_ = out.Close()
	if err := fs.Commit(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "100000,omega,zeta") {
		t.Fatalf("expected appended tail line")
	}
}

func TestProductionCodeAvoidsReadFile(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("txfs.go"))
	if err != nil {
		t.Fatalf("read txfs.go: %v", err)
	}
	if strings.Contains(string(data), "os.ReadFile(") {
		t.Fatalf("txfs.go must avoid os.ReadFile to keep commit path streaming")
	}
}
