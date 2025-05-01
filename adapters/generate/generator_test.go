package generate

import (
	"context"
	"strings"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/mock"
)

func TestGenerateDockerfile(t *testing.T) {
	tests := []struct {
		name     string
		facts    core.Facts
		wantErr  bool
		checkers []func(string) bool
	}{
		{
			name: "Spring Boot with distroless",
			facts: core.Facts{
				StackType: "spring-boot",
				BuildTool: "maven",
				Port:      8080,
			},
			wantErr: false,
			checkers: []func(string) bool{
				func(s string) bool {
					return strings.Contains(s, "gcr.io/distroless/java17-debian12")
				},
				func(s string) bool {
					return strings.Contains(s, "--mount=type=cache")
				},
			},
		},
		// ... existing test cases ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock LLM client
			mockLLM := mock.NewMockClient()
			generator := NewLLM(mockLLM)

			got, err := generator.Generate(context.Background(), tt.facts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				for _, checker := range tt.checkers {
					if !checker(got) {
						t.Errorf("Generated Dockerfile did not meet requirements")
					}
				}
			}
		})
	}
}
