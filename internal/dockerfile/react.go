package dockerfile

import (
	"fmt"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

// GenReact generates a Dockerfile for React projects
func GenReact(f types.Facts) string {
	var sb strings.Builder

	// Builder stage
	sb.WriteString("FROM node:20-bookworm as builder\n\n")
	sb.WriteString("WORKDIR /app\n\n")
	sb.WriteString("COPY package*.json ./\n")
	sb.WriteString("RUN npm ci --silent\n\n")
	sb.WriteString("COPY . .\n")
	sb.WriteString("RUN npm run build\n\n")

	// Runtime stage - use node if serve is specified, otherwise nginx
	if f.StartCmd != "" && strings.Contains(f.StartCmd, "serve") {
		sb.WriteString("FROM node:20-bookworm-slim\n\n")
		sb.WriteString("WORKDIR /app\n\n")
		sb.WriteString("# Copy package files and install serve\n")
		sb.WriteString("COPY package*.json ./\n")
		sb.WriteString("RUN npm ci --omit=dev --silent\n\n")
		sb.WriteString("# Copy built assets\n")
		if f.Artifact != "" {
			sb.WriteString(fmt.Sprintf("COPY --from=builder /app/%s ./%s\n", f.Artifact, f.Artifact))
		} else {
			sb.WriteString("COPY --from=builder /app/build ./build\n")
		}
		sb.WriteString("\nEXPOSE 80\n")
		sb.WriteString("ENV PORT=80\n\n")
		sb.WriteString("# Start serve\n")
		if f.StartCmd != "" {
			sb.WriteString(fmt.Sprintf("CMD %s\n", f.StartCmd))
		} else {
			sb.WriteString("CMD [\"npx\", \"serve\", \"-s\", \"build\"]\n")
		}
	} else {
		sb.WriteString("FROM nginx:alpine\n\n")
		sb.WriteString("# Copy built assets\n")
		if f.Artifact != "" {
			sb.WriteString(fmt.Sprintf("COPY --from=builder /app/%s /usr/share/nginx/html/\n", f.Artifact))
		} else {
			// Default to dist (Vite) or build (CRA)
			sb.WriteString("COPY --from=builder /app/dist /usr/share/nginx/html/ 2>/dev/null || COPY --from=builder /app/build /usr/share/nginx/html/\n")
		}

		// Add nginx config for SPA
		sb.WriteString("\n# Configure nginx for SPA\n")
		sb.WriteString("RUN echo $'\n")
		sb.WriteString("server {\n")
		sb.WriteString("    listen 80;\n")
		sb.WriteString("    location / {\n")
		sb.WriteString("        root /usr/share/nginx/html;\n")
		sb.WriteString("        try_files $uri $uri/ /index.html;\n")
		sb.WriteString("    }\n")
		sb.WriteString("}\n")
		sb.WriteString("' > /etc/nginx/conf.d/default.conf\n\n")

		// Expose port and add healthcheck
		sb.WriteString("EXPOSE 80\n")
		sb.WriteString("HEALTHCHECK --interval=30s --timeout=3s \\\n")
		sb.WriteString("  CMD curl -f http://localhost/ || exit 1\n")
	}

	return sb.String()
}
