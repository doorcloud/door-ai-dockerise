package core

import (
	"context"
	"io/fs"
)

// StackInfo represents information about a detected stack
type StackInfo struct {
	Name          string
	BuildTool     string
	Version       string
	SpecProvided  bool
	Confidence    float64
	DetectedFiles []string
	Port          int
}

// Fact represents a fact about a stack
type Fact struct {
	Key   string
	Value string
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LogSink is an interface for logging messages
type LogSink interface {
	Log(msg string)
}

// Generator generates a Dockerfile for a given stack
type Generator interface {
	Generate(ctx context.Context, facts Facts) (string, error)
	Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error)
}

// Verifier verifies that a generated file is valid
type Verifier interface {
	Verify(ctx context.Context, root string, generatedFile string) error
}

// ChatCompletion handles LLM interactions
type ChatCompletion interface {
	Complete(ctx context.Context, messages []Message) (string, error)
	GatherFacts(ctx context.Context, fsys fs.FS, stack StackInfo) (Facts, error)
	GenerateDockerfile(ctx context.Context, facts Facts) (string, error)
}

// FactProvider provides facts about a stack
type FactProvider interface {
	Facts(ctx context.Context, stack StackInfo) ([]Fact, error)
}

// DetectorChain implements Detector by trying each detector in sequence
type DetectorChain []Detector

// Detect implements the Detector interface for DetectorChain
func (c DetectorChain) Detect(ctx context.Context, fsys fs.FS, logSink LogSink) (StackInfo, bool, error) {
	for _, d := range c {
		info, found, err := d.Detect(ctx, fsys, logSink)
		if err != nil {
			return StackInfo{}, false, err
		}
		if found {
			return info, true, nil
		}
	}
	return StackInfo{}, false, nil
}

// Name returns the detector chain name
func (c DetectorChain) Name() string {
	return "chain"
}

// Describe returns a description of the detector chain
func (c DetectorChain) Describe() string {
	return "Tries each detector in sequence until one matches"
}

// SetLogSink sets the log sink for all detectors in the chain
func (c DetectorChain) SetLogSink(logSink LogSink) {
	for _, d := range c {
		if chain, ok := d.(DetectorChain); ok {
			chain.SetLogSink(logSink)
		}
	}
}

// Facts represents facts about a stack
type Facts struct {
	StackType string
	BuildTool string
	Port      int
	// Add more fields as needed
}

// Fixer provides a method to fix a Dockerfile
type Fixer interface {
	Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error)
}

// DockerfileGen generates a Dockerfile
type DockerfileGen interface {
	Generator
}
