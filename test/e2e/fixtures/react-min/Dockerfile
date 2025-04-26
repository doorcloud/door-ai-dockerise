# Build
FROM node:20-bookworm as build
WORKDIR /app
COPY . .
RUN npm ci && npm run build

# Run
FROM nginx:1.25-alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK CMD wget -qO- http://localhost || exit 1
