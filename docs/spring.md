# Spring Boot Detection and Fact Extraction

## Detection Rules

The Spring Boot detector (`SpringBootDetectorV2`) identifies Spring Boot projects by looking for specific build files and dependencies:

### Maven Projects
- Single-module: Checks for `pom.xml` in the root directory
- Multi-module: Checks for `pom.xml` in subdirectories

### Gradle Projects
- Groovy: Checks for `build.gradle` in the root directory
- Kotlin: Checks for `build.gradle.kts` in the root directory
- Multi-module: Checks for build files in subdirectories

The detector stops at the first positive hit and returns the following information:
- Name: "spring-boot"
- Build Tool: "maven" or "gradle"
- Detected Files: List of build files found

## Fact Extraction

The Spring Boot fact extractor (`Extractor`) gathers the following configuration from build files:

### Spec Keys
- `build_tool`: The build tool used ("maven" or "gradle")
- `jdk_version`: The Java version used for compilation
- `spring_boot_version`: The Spring Boot version
- `build_cmd`: The command to build the project
- `artifact`: The path to the built artifact
- `health_endpoint`: The health check endpoint (default: "/actuator/health")
- `ports`: The ports exposed by the application (default: [8080])

### Maven Extraction
- JDK Version: Extracted from `<java.version>` property
- Spring Boot Version: Extracted from `<spring-boot.version>` property
- Build Command: "mvn clean package -DskipTests"
- Artifact: "target/*.jar"

### Gradle Extraction
- JDK Version: Extracted from `sourceCompatibility` property
- Spring Boot Version: Extracted from `springBootVersion` property
- Build Command: "./gradlew build -x test"
- Artifact: "build/libs/*.jar"

## Health Check Behavior

After building the Docker image, the system performs a health check to verify that the application is running correctly:

1. The container is started with a random port mapped to the application port
2. Container logs are streamed to the output
3. The health endpoint is polled until:
   - A 200 OK response is received (success)
   - The timeout (30 seconds) is reached (failure)
4. The container is stopped and removed after the check

### Log Output
The health check process produces the following log messages:
```
docker run │ <container id>
health │ OK in <seconds> s
```

If the health check fails, the logs are included in the retry prompt to help diagnose the issue. 

# Spring Boot Support

## Overview
This document describes the Spring Boot support in the Dockerfile generator.

## Health Check Behavior
The generator performs the following health check steps:

1. Starts a container with the generated Dockerfile
2. Streams container logs
3. Polls the health endpoint (default: `/actuator/health`) until:
   - The endpoint returns HTTP 200 OK
   - A timeout occurs (default: 30 seconds)

The health check output is streamed in the following format:
```
docker run   │ <container-id>
health       │ OK in <seconds> s
```

If the health check fails, the generator will:
1. Stop and remove the container
2. Return an error with the container logs
3. Suggest fixes for the Dockerfile

## Configuration
The health check behavior can be configured through the following environment variables:
- `SPRING_HEALTH_ENDPOINT`: Custom health endpoint path (default: `/actuator/health`)
- `SPRING_HEALTH_TIMEOUT`: Health check timeout in seconds (default: 30)

## Example
```bash
# Run with custom health endpoint
SPRING_HEALTH_ENDPOINT=/health ./dockerfile-gen spring-boot-app

# Run with custom timeout
SPRING_HEALTH_TIMEOUT=60 ./dockerfile-gen spring-boot-app
```

# Spring Boot Detection Rules

## Build Tool Detection

### Maven
- Looks for `pom.xml` in the root directory
- Supports multi-module projects with parent POM

### Gradle
- Looks for `build.gradle` or `build.gradle.kts` in the root directory
- Supports multi-module projects with `include` or `include()`

## Fact Extraction

### JDK Version
1. Primary sources:
   - Maven: `<java.version>` property in `pom.xml`
   - Gradle: `sourceCompatibility` in build file
2. Fallback sources:
   - Maven: `<release>` tag in `.mvn/toolchains.xml`
   - Gradle: `languageVersion` in toolchain block of `build.gradle.kts`

### Spring Boot Version
1. Maven sources (in order):
   - `<spring-boot.version>` property
   - Parent POM version if groupId is `org.springframework.boot`
   - `spring-boot-dependencies` version in dependencyManagement
2. Gradle sources (in order):
   - Direct version properties
   - `libs.versions.toml` file

### Build Command
1. Wrapper detection:
   - Maven: Uses `./mvnw` if present, otherwise `mvn`
   - Gradle: Uses `./gradlew` if present, otherwise `gradle`
2. Multi-module support:
   - Maven: Adds `-pl <module> -am` for subdirectory artifacts
   - Gradle: Uses `:<module>:build` for subdirectory artifacts

### Artifact Path
1. Maven:
   - Default: `target/*.jar`
   - Rejects WAR packaging
2. Gradle:
   - Default: `build/libs/*.jar`
   - For multi-module: `<module>/build/libs/*.jar`

### Port and Health
1. Port detection:
   - Default: 8080
   - Overridden by `server.port` in application properties
2. Health endpoint:
   - Default: `/` (no actuator)
   - With actuator: `/actuator/health`
   - Customized by `management.endpoints.web.base-path`

### SBOM Path
1. Maven:
   - `target/bom.cdx.json` or `target/*.cdx.json` from CycloneDX plugin
2. Gradle:
   - `build/reports/bom.cdx.json` from CycloneDX task 