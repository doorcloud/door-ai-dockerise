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

1. **Search Depth**: Searches up to 4 levels deep in the project structure
2. **Maven Projects**:
   - Checks for parent pom.xml with Spring Boot parent
   - Checks for any springframework starter dependencies
   - Version fallback: If no version in current pom.xml, walks up parent POMs (max 2 levels)
3. **Gradle Projects**:
   - Matches any line containing "spring" and "boot" (case-insensitive) in:
     - build.gradle*
     - settings.gradle*
   - Supports both Groovy and Kotlin DSL

### Confidence Score

The detector provides a confidence score based on the number of signals found:

- 1 signal ⇒ 0.5
- 2 signals ⇒ 0.7
- 3 signals ⇒ 0.9
- ≥ 4 signals ⇒ 1.0

Signals include:
1. Spring Boot parent or platform BOM in Maven, or Spring Boot plugin in Gradle
2. Spring Boot starter dependencies
3. Spring Boot annotations in Java/Kotlin files
4. Spring Boot configuration files (application.properties/yml)
5. Spring Boot Maven plugin (Maven only)

### Version Detection

The detector extracts the Spring Boot version from:
- Maven: parent version in `pom.xml` or parent POMs (up to 2 levels up)
- Gradle: plugin version in `build.gradle`

Version strings are normalized by stripping everything after the first "-" (e.g. `3.2.0-SNAPSHOT` → `3.2.0`).

### Build Tool Preference

If a repository contains both Maven and Gradle build files, the detector will prefer Maven and skip the Gradle scan.

### Recursive Search

The detector recursively searches for build files up to a depth of 4 directories, allowing it to find Spring Boot modules in subdirectories like `services/api/pom.xml` or `apps/payment/build.gradle`.

### Default Port

If no explicit port is found in configuration files, the detector will use the default port 8080.

## Performance

The Spring Boot detector is optimized for performance:
- Uses regexp for efficient pattern matching
- Stops scanning Java/Kotlin files after finding first `@SpringBootApplication`
- Limits parent POM traversal to 2 levels
- Caches build file content for reuse

Benchmark results (M-class Apple Silicon):
- Average detection time: < 2ms per project
- Memory allocations: minimal
- Supports concurrent detection

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
    spring.NewSpringBootDetectorV3(),
    react.NewReactDetector(),
    node.NewNodeDetector(),
    mydetector.NewMyDetector(), // Add your detector here
}
```

## Testing

Each detector should have its own test suite. Additionally, integration tests in `integration_test.go` verify that all detectors work together correctly. 