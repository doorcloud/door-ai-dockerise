package mock

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type MockGenerator struct {
	dockerfile string
	err        error
}

func NewMockGenerator(dockerfile string, err error) *MockGenerator {
	return &MockGenerator{
		dockerfile: dockerfile,
		err:        err,
	}
}

func (g *MockGenerator) Generate(ctx context.Context, facts core.Facts) (string, error) {
	return g.dockerfile, g.err
}

func (g *MockGenerator) Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error) {
	return prevDockerfile, nil
}
