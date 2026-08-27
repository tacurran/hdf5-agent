package api

import (
	"strings"
	"testing"
)

func TestOpenAPIYAMLEmbedded(t *testing.T) {
	if len(OpenAPIYAML) == 0 {
		t.Fatal("empty spec")
	}
	if !strings.Contains(string(OpenAPIYAML), "openapi: 3.0.3") {
		t.Fatal("missing openapi header")
	}
	if !strings.Contains(string(OpenAPIYAML), "/api/v1/files") {
		t.Fatal("missing files path")
	}
}
