package llm

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/doorcloud/door-ai-dockerise/internal/facts"
)

// GenerateDockerfile generates a Dockerfile based on the provided facts.
func GenerateDockerfile(ctx context.Context, facts facts.Facts) (string, error) {
	// Convert facts to map for template
	factsMap := facts.ToMap()

	// Add build directory to template data
	if facts.BuildDir != "" {
		factsMap["build_dir"] = facts.BuildDir
	} else {
		factsMap["build_dir"] = "."
	}

	// Generate Dockerfile template
	tmpl := `FROM eclipse-temurin:17-jdk

WORKDIR /workspace

{{if ne .build_dir "."}}
COPY {{.build_dir}}/pom.xml .
COPY {{.build_dir}}/.mvn .mvn
COPY {{.build_dir}}/src src
{{else}}
COPY . .
{{end}}

RUN if [ -f "./mvnw" ]; then \
      chmod +x ./mvnw && \
      ./mvnw clean install; \
    else \
      curl -sL https://archive.apache.org/dist/maven/maven-3/3.9.0/binaries/apache-maven-3.9.0-bin.tar.gz | tar xz -C /tmp && \
      chmod +x /tmp/apache-maven-3.9.0/bin/mvn && \
      ln -s /tmp/apache-maven-3.9.0/bin/mvn /usr/bin/mvn && \
      mvn clean install; \
    fi

EXPOSE 8080

CMD ["java", "-jar", "target/*.jar"]`

	// Execute template
	var buf bytes.Buffer
	t := template.Must(template.New("dockerfile").Parse(tmpl))
	if err := t.Execute(&buf, factsMap); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
