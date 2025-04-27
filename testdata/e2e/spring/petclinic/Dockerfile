Here is a basic Dockerfile for a Spring Boot application using Maven.

```Dockerfile
# Start with a base image containing Java runtime (Here we are using OpenJDK)
FROM openjdk:8-jdk-alpine

# Add a volume pointing to /tmp (Optional)
VOLUME /tmp

# The application's jar file
ARG JAR_FILE=target/my_app.jar

# Copy the application's jar to the container
COPY ${JAR_FILE} app.jar

# Specify the start command
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]
```
- This Dockerfile assumes that you have built your Spring Boot application (also named "my_app") with Maven outside your Docker image and you got a JAR file at "target/my_app.jar".

If you would like to include the Maven building process in your Docker image, then you may use a multi-stage build Dockerfile:

```Dockerfile
# First stage: Build stage
FROM maven:3.8.1-openjdk-11-slim AS build

# Working directory
WORKDIR /build

# Copy pom.xml and src directory to the container
COPY pom.xml .
COPY src ./src

# Package the application
RUN mvn clean package

# Second stage: Run stage
FROM openjdk:11-jre-slim

# Application's port
EXPOSE 8080

# Working directory
WORKDIR /app

# Copy JAR file from the build stage
COPY --from=build /build/target/*.jar ./app.jar

# Start the application
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","./app.jar"]
```
- This Dockerfile builds the Spring Boot application from source directly inside the Docker image. All you have to prepare is your source code and pom.xml. 
- This Dockerfile also benefits from Docker's layer cache to speed up building process.
- Adjust the Java version and Maven version according to your requirements.
- Remember to replace the "*.jar" with your actual jar file name if it does not match the format.