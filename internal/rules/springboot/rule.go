package springboot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
	"github.com/doorcloud/door-ai-dockerise/pkg/rule"
)

// Rule implements the Rule interface for Spring Boot projects.
type Rule struct {
	logger    *slog.Logger
	llmClient *llm.Client
	config    *rules.RuleConfig
}

// NewRule creates a new Spring Boot rule.
func NewRule(logger *slog.Logger, llmClient *llm.Client, config *rules.RuleConfig) *Rule {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	return &Rule{
		logger:    logger,
		llmClient: llmClient,
		config:    config,
	}
}

// Detect determines if the given path contains a Spring Boot project.
func (r *Rule) Detect(path string) bool {
	stackRule, err := Detect(path)
	if err != nil {
		r.logger.Error("Failed to detect Spring Boot", "error", err)
		return false
	}
	return stackRule != nil
}

// Snippets returns relevant code snippets from the Spring Boot project.
func (r *Rule) Snippets(path string) ([]snippet.T, error) {
	pomFiles, err := r.findFiles(path, "pom.xml")
	if err != nil {
		return nil, fmt.Errorf("finding pom.xml: %w", err)
	}

	var snippets []snippet.T
	for _, pomFile := range pomFiles {
		snip, err := snippet.ReadFile(pomFile)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", pomFile, err)
		}
		snippets = append(snippets, snip)
	}

	return snippets, nil
}

// Facts extracts Spring Boot specific facts from the given snippets.
func (r *Rule) Facts(ctx context.Context, snippets []snippet.T, client facts.LLMClient) (facts.Facts, error) {
	var content []string
	for _, s := range snippets {
		content = append(content, s.Content)
	}

	factsMap, err := client.GenerateFacts(ctx, content)
	if err != nil {
		return facts.Facts{}, fmt.Errorf("extracting facts: %w", err)
	}

	f, err := facts.FromJSON(factsMap)
	if err != nil {
		return facts.Facts{}, fmt.Errorf("converting facts: %w", err)
	}

	return f, nil
}

// Dockerfile generates a Dockerfile for the project.
func (r *Rule) Dockerfile(ctx context.Context, f facts.Facts, llmClient facts.LLMClient) (string, error) {
	factsMap := f.ToMap()
	dockerfile, err := llmClient.GenerateDockerfile(ctx, factsMap)
	if err != nil {
		return "", fmt.Errorf("generating dockerfile: %w", err)
	}

	return dockerfile, nil
}

// Register the Spring Boot rule
func init() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := &rules.RuleConfig{}
	rule := NewRule(logger, nil, config)
	rules.Register(rule)
}

// findFiles finds all files matching the given patterns in the directory.
func (r *Rule) findFiles(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

type SpringBootRule struct {
	rule.BaseRule
}

func init() {
	rule.RegisterDefault("springboot", &SpringBootRule{
		BaseRule: rule.NewBaseRule(slog.Default()),
	})
}

func (r *SpringBootRule) Detect(path string) bool {
	// Check for pom.xml or build.gradle
	matches, err := doublestar.Glob(os.DirFS(path), "**/{pom.xml,build.gradle}")
	if err != nil {
		return false
	}
	if len(matches) == 0 {
		return false
	}

	// Check for Spring Boot dependencies
	buildFile := filepath.Join(path, matches[0])
	content, err := os.ReadFile(buildFile)
	if err != nil {
		return false
	}

	// Simple check for Spring Boot dependencies
	return strings.Contains(string(content), "spring-boot")
}

func (r *SpringBootRule) Snippets(path string) ([]snippet.T, error) {
	// TODO: Implement snippet extraction
	return nil, nil
}

func (r *SpringBootRule) Facts(ctx context.Context, snips []snippet.T, c *llm.Client) (facts.Facts, error) {
	return facts.Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: "maven",
		BuildCmd:  "mvn clean package -DskipTests",
		Artifact:  "target/*.jar",
		Ports:     []int{8080},
		Health:    "/actuator/health",
		BaseHint:  "eclipse-temurin:17-jdk",
	}, nil
}

func (r *SpringBootRule) Dockerfile(ctx context.Context, f facts.Facts, c *llm.Client) (string, error) {
	return `FROM eclipse-temurin:17-jdk AS builder
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn clean package -DskipTests

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
}
