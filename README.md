# dockergen (Spring-only)

A **tiny CLI** that:
1. Detects a Spring Boot repo  
2. Asks OpenAI for the best Dockerfile  
3. Writes it and runs `docker build`

```bash
export OPENAI_API_KEY=sk-…
dockergen --tag myapp:latest
docker run -p 8080:8080 myapp:latest
```

Flags `--repo`, `--tag`, `--retry`. Requires Go 1.22 and Docker. 