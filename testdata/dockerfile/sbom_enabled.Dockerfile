# Build stage
FROM eclipse-temurin:17-jdk as builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn -q package -DskipTests

# Runtime stage
FROM gcr.io/distroless/java17-debian12
WORKDIR /app
COPY --from=builder /app/target/*.jar /app/app.jar
COPY target/bom.cdx.json /app/sbom.cdx.json
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/app.jar"] 