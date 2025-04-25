package rules

import (
	"errors"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

// ErrUnknownStack is returned when no rule matches the project
var ErrUnknownStack = errors.New("unknown technology stack")

// DetectStack tries to detect the technology stack in the given repository
func DetectStack(fsys fs.FS) (*detect.RuleInfo, error) {
	rule, err := detect.Detect(fsys)
	if err != nil {
		return nil, ErrUnknownStack
	}
	return &rule, nil
}

// Registry implements the types.Registry interface
type Registry struct {
	detectors []types.Detector
}

// NewRegistry creates a new empty registry
func NewRegistry() *Registry {
	reg := &Registry{
		detectors: make([]types.Detector, 0),
	}

	// Register build plans
	reg.Register(&springDetector{})
	reg.Register(&nodeDetector{})

	return reg
}

// Register adds a detector to the registry
func (r *Registry) Register(detector types.Detector) {
	r.detectors = append(r.detectors, detector)
}

// GetDetectors returns all registered detectors
func (r *Registry) GetDetectors() []types.Detector {
	return r.detectors
}

// Detect tries each registered detector in order until one matches
func (r *Registry) Detect(fsys fs.FS) (detect.RuleInfo, bool) {
	for _, d := range r.detectors {
		detected, err := d.Detect(fsys)
		if err != nil {
			continue
		}
		if detected {
			return detect.RuleInfo{
				Name: d.Name(),
			}, true
		}
	}
	return detect.RuleInfo{}, false
}

// GetFacts extracts facts about the project using the given rule
func GetFacts(fsys fs.FS, rule detect.RuleInfo) (types.Facts, error) {
	switch rule.Name {
	case "spring-boot":
		if rule.Tool == "gradle" {
			return types.Facts{
				Language:  "java",
				Framework: "spring-boot",
				BuildTool: "gradle",
				BuildCmd:  "./gradlew bootJar -x test",
				BuildDir:  ".",
				StartCmd:  "java -jar build/libs/*.jar",
				Artifact:  "build/libs/*.jar",
				Ports:     []int{8080},
				Health:    "/actuator/health",
				BaseImage: "openjdk:11-jdk",
				Env:       map[string]string{},
			}, nil
		}
		return types.Facts{
			Language:  "java",
			Framework: "spring-boot",
			BuildTool: "maven",
			BuildCmd:  "./mvnw -q package -DskipTests",
			BuildDir:  ".",
			StartCmd:  "java -jar target/*.jar",
			Artifact:  "target/*.jar",
			Ports:     []int{8080},
			Health:    "/actuator/health",
			BaseImage: "openjdk:11-jdk",
			Env:       map[string]string{},
		}, nil
	case "node":
		if rule.Tool == "pnpm" {
			return types.Facts{
				Language:  "javascript",
				Framework: "node",
				BuildTool: "pnpm",
				BuildCmd:  "pnpm install --frozen-lockfile && pnpm run build",
				BuildDir:  ".",
				StartCmd:  "pnpm start",
				Artifact:  "dist/**",
				Ports:     []int{3000},
				Health:    "/health",
				BaseImage: "node:18-alpine",
				Env:       map[string]string{},
			}, nil
		}
		return types.Facts{
			Language:  "javascript",
			Framework: "node",
			BuildTool: "npm",
			BuildCmd:  "npm install",
			BuildDir:  ".",
			StartCmd:  "npm start",
			Artifact:  ".",
			Ports:     []int{3000},
			Health:    "/health",
			BaseImage: "node:18-alpine",
			Env:       map[string]string{},
		}, nil
	default:
		return types.Facts{}, nil
	}
}
