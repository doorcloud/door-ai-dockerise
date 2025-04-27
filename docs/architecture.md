# Architecture

This document describes the architecture of the Dockerfile Generator.

## Layer Diagram

```mermaid
graph TB
    subgraph "Entry Points"
        CLI[cmd/dockergen]
    end

    subgraph "Core Domain"
        Core[core]
        Core --> Interfaces[core/interfaces]
        Core --> Errors[core/errs]
        Core --> Logs[core/logs]
    end

    subgraph "Pipeline"
        Pipeline[pipeline]
        Pipeline --> Orchestrator[pipeline/orchestrator]
    end

    subgraph "Adapters"
        Detectors[adapters/rules/*]
        Facts[adapters/facts]
        Generator[adapters/generate]
        Verifiers[adapters/verifiers]
    end

    subgraph "Drivers"
        Docker[drivers/docker]
        K8sJob[drivers/k8sjob]
    end

    subgraph "Providers"
        LLM[providers/llm]
        LLM --> OpenAI[providers/llm/openai]
        LLM --> Ollama[providers/llm/ollama]
    end

    CLI --> Pipeline
    Pipeline --> Core
    Pipeline --> Adapters
    Adapters --> Core
    Adapters --> Drivers
    Adapters --> Providers
    Drivers --> Core
    Providers --> Core
```

## Component Responsibilities

### Core Domain
- **core**: Core domain types and interfaces
- **core/interfaces**: Interface definitions for all components
- **core/errs**: Error types and handling
- **core/logs**: Logging interfaces and utilities

### Pipeline
- **pipeline**: Main pipeline implementation
- **pipeline/orchestrator**: Orchestrates the detection, generation, and verification flow

### Adapters
- **adapters/rules/***: Stack detection rules (React, Spring Boot, etc.)
- **adapters/facts**: Fact gathering implementations
- **adapters/generate**: Dockerfile generation implementations
- **adapters/verifiers**: Dockerfile verification implementations

### Drivers
- **drivers/docker**: Docker API integration
- **drivers/k8sjob**: Kubernetes Job integration

### Providers
- **providers/llm**: LLM provider interface
- **providers/llm/openai**: OpenAI integration
- **providers/llm/ollama**: Ollama integration

## Data Flow

1. User invokes CLI
2. CLI creates and runs Pipeline
3. Pipeline uses Orchestrator to:
   - Detect stack using Detectors
   - Gather facts using FactProviders
   - Generate Dockerfile using Generator
   - Verify Dockerfile using Verifier
4. Results flow back to user

## Design Principles

1. **Clean Architecture**
   - Core domain is independent of external concerns
   - Dependencies point inward
   - Adapters translate between layers

2. **Interface Segregation**
   - Small, focused interfaces
   - Components depend only on what they need

3. **Dependency Inversion**
   - High-level modules don't depend on low-level modules
   - Both depend on abstractions

4. **Single Responsibility**
   - Each component has one reason to change
   - Clear separation of concerns

5. **Open/Closed**
   - Easy to extend without modification
   - New detectors/generators can be added without changing core 