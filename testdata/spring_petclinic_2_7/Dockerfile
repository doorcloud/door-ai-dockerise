# Build stage
FROM eclipse-temurin:17-jdk as builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn -q package -DskipTests
RUN java -Djarmode=layertools extract --destination layers --jar target/*.jar

# Runtime stage
FROM gcr.io/distroless/java17-debian12
WORKDIR /app
COPY --from=builder /app/layers/dependencies ./
COPY --from=builder /app/layers/spring-boot-loader ./
COPY --from=builder /app/layers/snapshot-dependencies ./
COPY --from=builder /app/layers/application ./
EXPOSE 8080
ENTRYPOINT ["java", "org.springframework.boot.loader.JarLauncher"] 