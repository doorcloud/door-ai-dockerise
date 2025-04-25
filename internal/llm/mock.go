package llm

import "context"

// Mock implements Client interface for testing
type Mock struct {
	FactsJSON  string
	Dockerfile string
}

// Chat implements Client.Chat by returning canned responses
func (m *Mock) Chat(_ context.Context, _ string) (string, error) {
	if m.FactsJSON != "" {
		out := m.FactsJSON
		m.FactsJSON = "" // next call returns Dockerfile
		return out, nil
	}
	return m.Dockerfile, nil
}
