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
		"extension/bus_language_core.js":         false,
		"extension/extension.js":                 false,
		"extension/language-server.js":           false,
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

func TestVSCodeBusLanguageGrammarCoversBusfileFixtures(t *testing.T) {
	t.Parallel()

	root := dispatcherModuleRoot(t)
	grammarPath := filepath.Join(root, "editors", "vscode-bus-language", "syntaxes", "bus.tmLanguage.json")
	cmd := exec.Command("python3", filepath.Join(root, "scripts", "check_vscode_bus_language_grammar.py"), grammarPath)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check grammar fixtures: %v\n%s", err, output)
	}
}

func TestVSCodeBusLanguageReleaseSurfaceReport(t *testing.T) {
	t.Parallel()

	root := dispatcherModuleRoot(t)
	cmd := exec.Command("python3", filepath.Join(root, "scripts", "check_vscode_bus_language_release.py"), "--format", "json")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check release surface: %v\n%s", err, output)
	}
	var report struct {
		ExtensionID      string   `json:"extension_id"`
		OpenVSXSlug      string   `json:"openvsx_slug"`
		Version          string   `json:"version"`
		VSIXPath         string   `json:"vsix_path"`
		VSIXRelease      string   `json:"vsix_release_asset"`
		SupportedEditors []string `json:"supported_editors"`
	}
	if err := json.Unmarshal(output, &report); err != nil {
		t.Fatalf("parse release report: %v\n%s", err, output)
	}
	if report.ExtensionID != "busdk.language-bus" {
		t.Fatalf("unexpected extension id %q", report.ExtensionID)
	}
	if report.OpenVSXSlug != "busdk/language-bus" {
		t.Fatalf("unexpected Open VSX slug %q", report.OpenVSXSlug)
	}
	if report.VSIXPath == "" || report.VSIXRelease == "" || report.Version == "" {
		t.Fatalf("unexpected empty release fields: %+v", report)
	}
	if len(report.SupportedEditors) == 0 {
		t.Fatalf("expected supported editors in report: %+v", report)
	}
}

func TestTreeSitterBusLanguageContract(t *testing.T) {
	t.Parallel()

	root := dispatcherModuleRoot(t)
	cmd := exec.Command("node", filepath.Join(root, "scripts", "check_tree_sitter_bus_language.js"))
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check tree-sitter contract: %v\n%s", err, output)
	}
}

func TestBusLanguageServerSemanticTokens(t *testing.T) {
	t.Parallel()

	root := dispatcherModuleRoot(t)
	cmd := exec.Command("python3", filepath.Join(root, "scripts", "check_bus_language_server.py"))
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check language server: %v\n%s", err, output)
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
