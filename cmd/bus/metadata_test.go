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
