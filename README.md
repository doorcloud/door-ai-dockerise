# dockergen (Spring-only)

Tiny CLI that:
1. Detects a Spring Boot repo
2. Asks OpenAI for the best Dockerfile
3. Writes it, then runs `docker build`

```bash
export OPENAI_API_KEY=sk-…
dockergen --tag myapp:latest
docker run -p 8080:8080 myapp:latest
```

Requires Go 1.22 and Docker. Flags: `--repo`, `--tag`, `--retry`. 