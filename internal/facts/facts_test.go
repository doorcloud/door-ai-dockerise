package facts

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

func TestInferWithClient(t *testing.T) {
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`<project></project>`),
		},
	}

	facts, err := InferWithClient(context.Background(), fsys, detect.Rule{
		Name: "spring-boot",
		Tool: "maven",
	}, &llm.MockClient{})
	if err != nil {
		t.Fatalf("InferWithClient() error = %v", err)
	}

	if facts.Language != "java" {
		t.Errorf("Expected language 'java', got '%s'", facts.Language)
	}
	if facts.Framework != "spring-boot" {
		t.Errorf("Expected framework 'spring-boot', got '%s'", facts.Framework)
	}
}
