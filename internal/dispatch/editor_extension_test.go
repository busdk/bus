package dispatch_test

import (
	"archive/zip"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVSCodeBusLanguageManifestRegistersBusFiles(t *testing.T) {
	t.Parallel()

	root := dispatcherModuleRoot(t)
	packagePath := filepath.Join(root, "editors", "vscode-bus-language", "package.json")
	raw, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	var manifest struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Contributes struct {
			Languages []struct {
				ID         string   `json:"id"`
				Extensions []string `json:"extensions"`
			} `json:"languages"`
		} `json:"contributes"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse package.json: %v", err)
	}
	if manifest.Name != "language-bus" {
		t.Fatalf("unexpected extension name %q", manifest.Name)
	}
	if manifest.DisplayName == "" {
		t.Fatal("displayName should not be empty")
	}
	foundBus := false
	for _, language := range manifest.Contributes.Languages {
		if language.ID != "bus" {
			continue
		}
		for _, extension := range language.Extensions {
			if extension == ".bus" {
				foundBus = true
			}
		}
	}
	if !foundBus {
		t.Fatal("expected .bus extension registration in package.json")
	}
}

func TestVSCodeBusLanguagePackagerProducesVSIX(t *testing.T) {
	t.Parallel()

	root := dispatcherModuleRoot(t)
	outputPath := filepath.Join(t.TempDir(), "bus-language.vsix")
	cmd := exec.Command("python3", filepath.Join(root, "scripts", "package_vscode_bus_language.py"), "--output", outputPath)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("package extension: %v\n%s", err, output)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("stat packaged vsix: %v", err)
	}
	archive, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("open vsix: %v", err)
	}
	defer func() { _ = archive.Close() }()
	requiredEntries := map[string]bool{
		"[Content_Types].xml":                    false,
		"extension.vsixmanifest":                 false,
		"extension/package.json":                 false,
		"extension/README.md":                    false,
		"extension/LICENSE.md":                   false,
		"extension/language-configuration.json":  false,
		"extension/syntaxes/bus.tmLanguage.json": false,
	}
	for _, file := range archive.File {
		if _, ok := requiredEntries[file.Name]; ok {
			requiredEntries[file.Name] = true
		}
	}
	for name, present := range requiredEntries {
		if !present {
			t.Fatalf("expected %s in packaged vsix", name)
		}
	}
}

func dispatcherModuleRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
