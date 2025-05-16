package prompt

import "fmt"

const tpl = `You are a DevOps expert. Write the complete Dockerfile
to build and run a Java Spring Boot project named %s.
Requirements:
- Use multi-stage build with Eclipse Temurin JDK 17.
- Copy the fat JAR to a slim runtime stage.
- Expose port 8080; run as non-root user app:app.
Return ONLY the Dockerfile content.`

// Render fills the template with the repo name.
func Render(repoName string) string { return fmt.Sprintf(tpl, repoName) }
