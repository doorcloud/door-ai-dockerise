package core

import (
	"context"
)

// StackInfo represents information about a detected stack
type StackInfo struct {
	Name string
	Meta map[string]string
}

// Fact represents a fact about a stack
type Fact struct {
	Key   string
	Value string
}

// Message represents a chat message
type Message struct {
	Role    string
	Content string
}

// Detector detects the type of application in a directory
type Detector interface {
	Detect(ctx context.Context, dir string) (StackInfo, error)
}

// Generator generates a Dockerfile for a given stack
type Generator interface {
	Generate(ctx context.Context, stack StackInfo, facts []Fact) (string, error)
}

// Verifier verifies that a generated file is valid
type Verifier interface {
	Verify(ctx context.Context, root string, generatedFile string) error
}

// ChatCompletion is responsible for chat-based completions
type ChatCompletion interface {
	Chat(ctx context.Context, msgs []Message) (Message, error)
}

// DetectorChain implements Detector by trying each detector in sequence
type DetectorChain []Detector

// Detect implements the Detector interface for DetectorChain
func (c DetectorChain) Detect(ctx context.Context, dir string) (StackInfo, error) {
	for _, d := range c {
		info, err := d.Detect(ctx, dir)
		if err != nil {
			return StackInfo{}, err
		}
		if info.Name != "" {
			return info, nil
		}
	}
	return StackInfo{}, nil
}
