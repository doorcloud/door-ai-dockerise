package llm

// mockClient implements the Client interface using local fixtures
type mockClient struct{}

func (c *mockClient) Chat(model, prompt string) (string, error) {
	if model == "facts" {
		return `{"language":"Java","framework":"Spring Boot"}`, nil
	}
	return "FROM eclipse-temurin:17-jdk\nWORKDIR /app\nCOPY . .\nRUN ./mvnw package\nCMD [\"java\", \"-jar\", \"target/*.jar\"]", nil
}
