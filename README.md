# Dockerfile Generator

A tool that automatically generates Dockerfiles for your applications by analyzing your codebase or using a provided specification.

## Quick Start

### Installation

```bash
go install github.com/doorcloud/door-ai-dockerise/cmd/dockergen@latest
```

### Usage

There are two ways to use the tool:

1. **Spec-First Approach** (Recommended for known stacks)
   ```bash
   # Generate Dockerfile using a stack specification
   dockergen --spec examples/spring-boot/stack.yaml
   ```

2. **Code-First Approach** (For automatic detection)
   ```bash
   # Generate Dockerfile by analyzing your codebase
   dockergen --path ./my-project
   ```

### Examples

#### Spring Boot Example

Create a `stack.yaml` file:
```yaml
language: java
framework: springboot
version: "3.2"
buildTool: maven
params:
  port: "8080"
  jdkVersion: "17"
  buildArgs: "-DskipTests"
  baseImage: "eclipse-temurin:17-jre-alpine"
```

Then run:
```bash
dockergen --spec stack.yaml
```

#### Node.js Example

For a Node.js project, simply point to your project directory:
```bash
dockergen --path ./my-node-app
```

## Features

- Automatic stack detection
- Support for multiple frameworks and languages
- Customizable through stack specifications
- Dockerfile verification
- Build and test the generated Dockerfile

## Supported Stacks

- Spring Boot
- Node.js
- React
- More coming soon...

## Development

### Building from Source

```bash
git clone https://github.com/doorcloud/door-ai-dockerise
cd door-ai-dockerise
go build -o dockergen ./cmd/dockergen
```

### Running Tests

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 