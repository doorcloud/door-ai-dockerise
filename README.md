# Dockerfile Generator

A Go-based tool for generating Dockerfiles with customizable configurations.

## Project Structure

```
.
├── adapters/     # Adapters for different components (detectors, generators, etc.)
├── cmd/          # Main application entry points
├── core/         # Core interfaces and types
├── drivers/      # External service drivers (Docker, etc.)
├── pipeline/     # Pipeline implementation (v2)
├── providers/    # External service providers (LLM, etc.)
├── pkg/          # Public library code
└── test/         # Test files
```

## Prerequisites

- Go 1.21 or later
- Docker (for testing)
- OpenAI API key (optional, for real API testing)

## Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd dockerfile-gen
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Copy the environment file:
   ```bash
   cp .env.example .env
   ```

4. Configure your `.env` file with appropriate settings.

## Building

To build the project:
```bash
go build ./cmd/...
```

## Testing

Run the test suite:
```bash
go test ./...
```

### Offline Testing
The test suite can run without an OpenAI API key by using mock responses:
```bash
DG_MOCK_LLM=1 go test ./...
```

To run tests with the real OpenAI API:
```bash
OPENAI_API_KEY=your_key go test ./...
```

### End-to-End Testing
To run the full end-to-end test suite:
```bash
DG_MOCK_LLM=1 DG_E2E=1 go test -tags=integration ./test/e2e/...
```

## Debugging

The following environment variables can be used to enable various debug features:

- `DEBUG=true` - Enable global verbose logging with file and line information
- `DG_DEBUG=1` - Enable additional logging in docker-gen specific code paths
- `OPENAI_LOG_LEVEL=debug` - Show raw HTTP traces for OpenAI API calls
- `DG_E2E=1` - Enable the full end-to-end test suite (longer running tests)
- `DG_MOCK_LLM=1` - Use mock LLM responses instead of real API calls

These can be set in your `.env` file or directly in the shell before running commands.

## Supported Stacks

The tool currently supports generating Dockerfiles for:

- React (Node.js-based)
- Spring Boot (Java-based)

Each stack is detected automatically based on project files and dependencies.

## Usage

1. Basic usage:
   ```bash
   ./cmd/dockergen/dockergen <project-path>
   ```

2. With custom configuration:
   ```bash
   OPENAI_API_KEY=your_key ./cmd/dockergen/dockergen <project-path>
   ```

The tool will:
1. Detect the project stack
2. Gather relevant facts about the project
3. Generate a Dockerfile
4. Verify the Dockerfile can be built

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[Add license information here] 