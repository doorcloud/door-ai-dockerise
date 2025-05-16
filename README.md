# dockergen (Spring-only)

A **tiny CLI** that:
1. Detects if the current repo is a Spring Boot project.
2. Asks OpenAI to generate an optimized Dockerfile.
3. Writes `Dockerfile` and runs `docker build`.

---

## Requirements
- **Go 1.22+** (only for building the CLI)
- **Docker** daemon running
- `OPENAI_API_KEY` exported in your shell

```bash
go build -o dockergen ./cmd/...
```

---

## Quick start

```bash
# inside any Spring Boot repo
export OPENAI_API_KEY=sk-...
dockergen --tag myapp:latest  # creates Dockerfile + builds image
docker run -p 8080:8080 myapp:latest
```

### Flags

| Flag   | Default            | Description                      |
|--------|-------------------|----------------------------------|
| `--repo` | `.`              | Path to the source repository    |
| `--tag`  | `spring-app:latest` | Docker image tag               |
| `--retry` | `0`              | Times to re-ask the LLM on failure |

---

## How it works (3 steps)

1. **Detect**: looks for `pom.xml` or `build.gradle` to confirm Spring.
2. **Prompt**: fills a single template and calls ChatGPT (`gpt-4o-mini`).
3. **Build**: saves the Dockerfile and runs `docker build`.

That's it—no scans, no extra plugins, just fast Dockerization for Spring Boot. 