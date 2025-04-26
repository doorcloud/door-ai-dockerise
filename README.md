# Dockerfile Generator

A tool to automatically generate Dockerfiles for your projects using AI.

## Installation

```bash
go install github.com/doorcloud/door-ai-dockerise/cmd/dockergen@latest
```

## Usage

### Basic Usage

Generate a Dockerfile for a project in the current directory:

```bash
dockergen
```

### Advanced Usage

Generate a Dockerfile for a specific project directory:

```bash
dockergen --path ./repo
```

Generate a Dockerfile using a specific LLM provider:

```bash
# Using Ollama
dockergen --path ./repo --llm ollama

# Using OpenAI
dockergen --spec stack.yaml --llm openai --verbose
```

### Options

- `--path`: Path to the project directory (default: ".")
- `--spec`: Path to stack specification file (yaml/json)
- `--llm`: LLM provider to use (openai|ollama) (default: "openai")
- `--verbose`: Enable verbose logging
- `--debug`: Enable debug logging

## Configuration

### OpenAI

To use OpenAI, set the `OPENAI_API_KEY` environment variable:

```bash
export OPENAI_API_KEY=your-api-key
```

### Ollama

Ollama should be running locally on the default port (11434).

## Development

### Building

```bash
go build -o dockergen ./cmd/dockergen
```

### Testing

```bash
go test ./...
```

## License

MIT 