package txfs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkIsDeleted(b *testing.B) {
	for _, deletes := range []int{1, 10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("deletes_%d", deletes), func(b *testing.B) {
			fs := &FS{
				changes:   map[string]changeKind{},
				tombstone: make(map[string]struct{}, deletes),
			}
			for i := 0; i < deletes; i++ {
				fs.tombstone[fmt.Sprintf("dir_%05d", i)] = struct{}{}
			}
			target := fmt.Sprintf("dir_%05d/file.txt", deletes-1)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if !fs.isDeleted(target) {
					b.Fatalf("expected target to be tombstoned")
				}
			}
		})
	}
}

func BenchmarkMarkDelete(b *testing.B) {
	for _, existing := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("existing_%d", existing), func(b *testing.B) {
			fs := &FS{
				changes:   make(map[string]changeKind, existing+1),
				tombstone: make(map[string]struct{}, 1),
			}
			for j := 0; j < existing; j++ {
				fs.changes[fmt.Sprintf("dir_%05d/file.txt", j)] = changeReplace
			}
			// Warm once so repeated calls benchmark steady-state scan cost.
			fs.markDelete("dir_99999")

			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				fs.markDelete("dir_99999")
			}
		})
	}
}

func BenchmarkOpenFileRepeatedExistingPath(b *testing.B) {
	for _, rel := range []string{
		"file.txt",
		filepath.Join("deep", "nested", "tree", "path", "file.txt"),
	} {
		b.Run(rel, func(b *testing.B) {
			root := b.TempDir()
			overlay := filepath.Join(b.TempDir(), "overlay")
			fs, err := New(root, overlay)
			if err != nil {
				b.Fatalf("new fs: %v", err)
			}

			seed, err := fs.OpenFile(rel, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				b.Fatalf("seed open: %v", err)
			}
			if err := seed.Close(); err != nil {
				b.Fatalf("seed close: %v", err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, err := fs.OpenFile(rel, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err != nil {
					b.Fatalf("open: %v", err)
				}
				if err := f.Close(); err != nil {
					b.Fatalf("close: %v", err)
				}
			}
		})
	}
}

func BenchmarkOpenReadExistingPath(b *testing.B) {
	for _, rel := range []string{
		"file.txt",
		filepath.Join("deep", "nested", "tree", "path", "file.txt"),
	} {
		b.Run(rel, func(b *testing.B) {
			root := b.TempDir()
			overlay := filepath.Join(b.TempDir(), "overlay")
			basePath := filepath.Join(root, rel)
			if err := os.MkdirAll(filepath.Dir(basePath), 0o755); err != nil {
				b.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(basePath, []byte("value\n"), 0o644); err != nil {
				b.Fatalf("write: %v", err)
			}

			fs, err := New(root, overlay)
			if err != nil {
				b.Fatalf("new fs: %v", err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, err := fs.Open(rel)
				if err != nil {
					b.Fatalf("open: %v", err)
				}
				if err := f.Close(); err != nil {
					b.Fatalf("close: %v", err)
				}
			}
		})
	}
}

func BenchmarkStatExistingPath(b *testing.B) {
	for _, rel := range []string{
		"file.txt",
		filepath.Join("deep", "nested", "tree", "path", "file.txt"),
	} {
		b.Run(rel, func(b *testing.B) {
			root := b.TempDir()
			overlay := filepath.Join(b.TempDir(), "overlay")
			basePath := filepath.Join(root, rel)
			if err := os.MkdirAll(filepath.Dir(basePath), 0o755); err != nil {
				b.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(basePath, []byte("value\n"), 0o644); err != nil {
				b.Fatalf("write: %v", err)
			}

			fs, err := New(root, overlay)
			if err != nil {
				b.Fatalf("new fs: %v", err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := fs.Stat(rel); err != nil {
					b.Fatalf("stat: %v", err)
				}
			}
		})
	}
}
