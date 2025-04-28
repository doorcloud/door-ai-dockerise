# Detectors Package

This package provides a unified interface for detecting project types and their configurations. The detectors are organized in a modular way to support easy addition of new project type detectors.

## Directory Layout

```
adapters/detectors/
├── README.md           # This file
├── registry.go         # Central orchestrator and registry
├── node/              # Node.js detector
├── react/             # React detector
└── spring/            # Spring Boot detector
```

## How It Works

The detector system uses a registry-based approach where all detectors are registered in a central location (`registry.go`). Each detector implements the `core.Detector` interface.

## Spring Boot Detection Rules

The Spring Boot detector uses the following heuristics:

1. **Search Depth**: Searches up to 2 levels deep in the project structure
2. **Maven Projects**:
   - Checks for parent pom.xml with Spring Boot parent
   - Checks for any springframework starter dependencies
3. **Gradle Projects**:
   - Matches any line containing "spring" and "boot" (case-insensitive) in:
     - build.gradle*
     - settings.gradle*
   - Supports both Groovy and Kotlin DSL

## Adding a New Detector

To add a new detector:

1. Create a new directory under `adapters/detectors/` for your detector
2. Implement the `core.Detector` interface
3. Add your detector to the registry in `registry.go`

Example:

```go
// adapters/detectors/mydetector/detector.go
package mydetector

type MyDetector struct {
    logSink core.LogSink
}

func NewMyDetector() *MyDetector {
    return &MyDetector{}
}

func (d *MyDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
    // Implementation
}

func (d *MyDetector) Name() string {
    return "my-detector"
}

func (d *MyDetector) SetLogSink(logSink core.LogSink) {
    d.logSink = logSink
}
```

Then add it to the registry in `registry.go`:

```go
var registry = []core.Detector{
    spring.NewSpringBootDetectorV2(),
    react.NewReactDetector(),
    node.NewNodeDetector(),
    mydetector.NewMyDetector(), // Add your detector here
}
```

## Testing

Each detector should have its own test suite. Additionally, integration tests in `integration_test.go` verify that all detectors work together correctly. 