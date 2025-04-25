FROM eclipse-temurin:17-jdk
WORKDIR /app
COPY . .
RUN ./mvnw package
CMD ["java", "-jar", "target/*.jar"]