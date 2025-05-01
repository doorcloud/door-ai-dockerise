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
		{
			name: "Spring Boot 2.7 with layered JAR",
			facts: core.Facts{
				StackType: "spring-boot",
				BuildTool: "maven",
				Port:      8080,
			},
			wantErr: false,
			checkers: []func(string) bool{
				func(s string) bool {
					return strings.Contains(s, "java -Djarmode=layertools extract")
				},
				func(s string) bool {
					return strings.Contains(s, "COPY --from=builder /app/layers/dependencies ./")
				},
				func(s string) bool {
					return strings.Contains(s, "COPY --from=builder /app/layers/spring-boot-loader ./")
				},
				func(s string) bool {
					return strings.Contains(s, "COPY --from=builder /app/layers/snapshot-dependencies ./")
				},
				func(s string) bool {
					return strings.Contains(s, "COPY --from=builder /app/layers/application ./")
				},
				func(s string) bool {
					return strings.Contains(s, "ENTRYPOINT [\"java\", \"org.springframework.boot.loader.JarLauncher\"]")
				},
			},
		},
		{
			name: "Spring Boot with SBOM",
			facts: core.Facts{
				StackType: "spring-boot",
				BuildTool: "maven",
				Port:      8080,
				SBOMPath:  "target/bom.cdx.json",
			},
			wantErr: false,
			checkers: []func(string) bool{
				func(s string) bool {
					return strings.Contains(s, "COPY target/bom.cdx.json /app/sbom.cdx.json")
				},
			},
		},
		// ... existing test cases ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock.NewMockClient()
			if tt.name == "Spring Boot 2.7 with layered JAR" {
				mockClient.SetResponse("spring-boot:maven", `# Build stage
FROM eclipse-temurin:17-jdk as builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn -q package -DskipTests
RUN java -Djarmode=layertools extract --destination layers --jar target/*.jar

# Runtime stage
FROM gcr.io/distroless/java17-debian12
WORKDIR /app
COPY --from=builder /app/layers/dependencies ./
COPY --from=builder /app/layers/spring-boot-loader ./
COPY --from=builder /app/layers/snapshot-dependencies ./
COPY --from=builder /app/layers/application ./
EXPOSE 8080
ENTRYPOINT ["java", "org.springframework.boot.loader.JarLauncher"]`)
			}
			generator := NewLLM(mockClient)
			got, err := generator.Generate(context.Background(), tt.facts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, checker := range tt.checkers {
				if !checker(got) {
					t.Errorf("Generate() checker %d failed", i)
				}
			}
		})
	}
}
