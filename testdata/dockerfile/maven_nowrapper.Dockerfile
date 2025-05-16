FROM maven:3.9-eclipse-temurin17 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn clean package

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"] 