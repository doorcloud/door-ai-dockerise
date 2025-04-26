# build stage
FROM node:20-alpine AS build
WORKDIR /app
COPY . .
RUN npm ci && npm run build

# runtime
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80 