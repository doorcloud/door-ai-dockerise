# Build stage
FROM eclipse-temurin:17-jdk as builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn -q package -DskipTests

# Runtime stage
FROM gcr.io/distroless/java17-debian12
WORKDIR /app
COPY --from=builder /app/target/*.jar /app/app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/app.jar"] 