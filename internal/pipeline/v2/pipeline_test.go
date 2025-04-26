package v2

import (
	"context"
	"errors"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type mockDetector struct {
	stack core.StackInfo
	err   error
}

func (m *mockDetector) Detect(ctx context.Context, path string) (core.StackInfo, error) {
	return m.stack, m.err
}

type mockFactProvider struct {
	facts []string
	err   error
}

func (m *mockFactProvider) Facts(ctx context.Context, stack core.StackInfo) ([]string, error) {
	return m.facts, m.err
}

type mockGenerator struct {
	dockerfile string
	err        error
}

func (m *mockGenerator) Generate(ctx context.Context, facts []string) (string, error) {
	return m.dockerfile, m.err
}

type mockVerifier struct {
	err error
}

func (m *mockVerifier) Verify(ctx context.Context, repoPath, dockerfile string) error {
	return m.err
}

func TestPipeline_Run(t *testing.T) {
	tests := []struct {
		name           string
		detector       core.Detector
		factProvider   core.FactProvider
		generator      core.DockerfileGen
		verifier       core.Verifier
		wantDockerfile string
		wantErr        bool
	}{
		{
			name: "successful run",
			detector: &mockDetector{
				stack: core.StackInfo{Name: "react"},
			},
			factProvider: &mockFactProvider{
				facts: []string{"language:javascript", "framework:react"},
			},
			generator: &mockGenerator{
				dockerfile: "FROM node:18\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD [\"npm\", \"start\"]",
			},
			verifier:       &mockVerifier{},
			wantDockerfile: "FROM node:18\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD [\"npm\", \"start\"]",
			wantErr:        false,
		},
		{
			name: "detection failure",
			detector: &mockDetector{
				err: errors.New("detection failed"),
			},
			factProvider:   &mockFactProvider{},
			generator:      &mockGenerator{},
			verifier:       &mockVerifier{},
			wantDockerfile: "",
			wantErr:        true,
		},
		{
			name: "facts collection failure",
			detector: &mockDetector{
				stack: core.StackInfo{Name: "react"},
			},
			factProvider: &mockFactProvider{
				err: errors.New("facts collection failed"),
			},
			generator:      &mockGenerator{},
			verifier:       &mockVerifier{},
			wantDockerfile: "",
			wantErr:        true,
		},
		{
			name: "generation failure",
			detector: &mockDetector{
				stack: core.StackInfo{Name: "react"},
			},
			factProvider: &mockFactProvider{
				facts: []string{"language:javascript", "framework:react"},
			},
			generator: &mockGenerator{
				err: errors.New("generation failed"),
			},
			verifier:       &mockVerifier{},
			wantDockerfile: "",
			wantErr:        true,
		},
		{
			name: "verification failure",
			detector: &mockDetector{
				stack: core.StackInfo{Name: "react"},
			},
			factProvider: &mockFactProvider{
				facts: []string{"language:javascript", "framework:react"},
			},
			generator: &mockGenerator{
				dockerfile: "FROM node:18\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD [\"npm\", \"start\"]",
			},
			verifier: &mockVerifier{
				err: errors.New("verification failed"),
			},
			wantDockerfile: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPipeline(
				[]core.Detector{tt.detector},
				[]core.FactProvider{tt.factProvider},
				tt.generator,
				tt.verifier,
			)

			got, err := p.Run(context.Background(), "testdata")
			if (err != nil) != tt.wantErr {
				t.Errorf("Pipeline.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantDockerfile {
				t.Errorf("Pipeline.Run() = %v, want %v", got, tt.wantDockerfile)
			}
		})
	}
}
