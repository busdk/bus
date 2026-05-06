package dispatch

import (
	"bufio"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryHasNoPrivateBusModuleImports(t *testing.T) {
	root := repoRoot(t)
	skipDirs := map[string]bool{
		".git":   true,
		".make":  true,
		"bin":    true,
		"vendor": true,
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, spec := range file.Imports {
			importPath := strings.Trim(spec.Path.Value, `"`)
			if strings.HasPrefix(importPath, "github.com/busdk/bus-") && !strings.HasPrefix(importPath, "github.com/busdk/bus-help/") {
				t.Errorf("%s imports forbidden private BusDK module package %q", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk imports: %v", err)
	}
}

func TestRepositoryHasNoPrivateBusModuleBuildDependencies(t *testing.T) {
	root := repoRoot(t)

	for _, relPath := range []string{"go.mod", "Makefile.local"} {
		path := filepath.Join(root, relPath)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", relPath, err)
		}
		text := string(data)
		privateText := strings.ReplaceAll(text, "github.com/busdk/bus-help", "")
		if strings.Contains(privateText, "github.com/busdk/bus-") {
			t.Fatalf("%s must not reference private github.com/busdk/bus-* modules", relPath)
		}
		if strings.Contains(text, "../bus-") || strings.Contains(text, `..\bus-`) {
			t.Fatalf("%s must not reference sibling private bus-* modules", relPath)
		}
	}

	makefileLocalPath := filepath.Join(root, "Makefile.local")
	f, err := os.Open(makefileLocalPath)
	if err != nil {
		t.Fatalf("open Makefile.local: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "MODULE_SRC_DEPS :="):
			assertNoPrivateModuleDeps(t, "MODULE_SRC_DEPS", strings.TrimSpace(strings.TrimPrefix(line, "MODULE_SRC_DEPS :=")))
		case strings.HasPrefix(line, "MODULE_BIN_DEPS :="):
			assertNoPrivateModuleDeps(t, "MODULE_BIN_DEPS", strings.TrimSpace(strings.TrimPrefix(line, "MODULE_BIN_DEPS :=")))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan Makefile.local: %v", err)
	}
}

func assertNoPrivateModuleDeps(t *testing.T, name string, value string) {
	t.Helper()
	for _, field := range strings.Fields(value) {
		if strings.HasPrefix(field, "bus-") {
			t.Fatalf("%s must not reference private bus-* module dependency %q", name, field)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}
