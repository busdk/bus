package main

import (
	"encoding/json"
	"testing"
)

func TestMetadataDocumentIncludesEnvironmentMetadata(t *testing.T) {
	doc := metadataDocument()
	if doc.Info.Title != "bus" {
		t.Fatalf("unexpected title: %s", doc.Info.Title)
	}
	if _, ok := doc.Metadata["io.busdk.environment"]; !ok {
		t.Fatal("missing Bus environment metadata")
	}
	if _, err := json.Marshal(doc); err != nil {
		t.Fatalf("metadata document is not JSON serializable: %v", err)
	}
}

func TestMetadataDocumentIncludesDiagnosticOptions(t *testing.T) {
	doc := metadataDocument()
	if len(doc.Commands) == 0 {
		t.Fatal("missing command metadata")
	}
	options := map[string]bool{}
	for _, option := range doc.Commands[0].Options {
		options[option.Name] = true
	}
	for _, name := range []string{"--verbose", "--trace", "--quiet", "--perf"} {
		if !options[name] {
			t.Fatalf("missing diagnostic option %s", name)
		}
	}
}
