package core

import (
	"context"
)

// StackInfo represents information about a detected stack
type StackInfo struct {
	Name string
	Meta map[string]string
}

// Fact represents a piece of information about a stack
type Fact struct {
	Key   string
	Value string
}

// Message represents a chat message
type Message struct {
	Role    string
	Content string
}

// Detector is responsible for detecting technology stacks
type Detector interface {
	Detect(ctx context.Context, root string) (StackInfo, error)
}

// FactProvider is responsible for gathering facts about a detected stack
type FactProvider interface {
	Facts(ctx context.Context, info StackInfo) ([]Fact, error)
}

// DockerfileGen is responsible for generating Dockerfiles
type DockerfileGen interface {
	Generate(ctx context.Context, facts []Fact) (string, error)
}

// Verifier is responsible for verifying Dockerfiles
type Verifier interface {
	Verify(ctx context.Context, root string, dockerfile string) error
}

// ChatCompletion is responsible for chat-based completions
type ChatCompletion interface {
	Chat(ctx context.Context, msgs []Message) (Message, error)
}

// DetectorChain implements Detector by trying each detector in sequence
type DetectorChain []Detector

// Detect implements the Detector interface for DetectorChain
func (c DetectorChain) Detect(ctx context.Context, root string) (StackInfo, error) {
	for _, d := range c {
		info, err := d.Detect(ctx, root)
		if err != nil {
			return StackInfo{}, err
		}
		if info.Name != "" {
			return info, nil
		}
	}
	return StackInfo{}, nil
}
