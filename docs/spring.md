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