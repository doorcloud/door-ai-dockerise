package loop

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

func TestRun(t *testing.T) {
	// Create test filesystem
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>org.springframework.samples</groupId>
    <artifactId>spring-petclinic</artifactId>
    <version>3.2.0-SNAPSHOT</version>
</project>`),
		},
		"mvnw": &fstest.MapFile{
			Data: []byte("#!/bin/sh\n# Maven wrapper script"),
			Mode: 0755,
		},
		".mvn/wrapper/maven-wrapper.jar": &fstest.MapFile{
			Data: []byte("mock jar file"),
		},
		".mvn/wrapper/maven-wrapper.properties": &fstest.MapFile{
			Data: []byte("distributionUrl=https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.4/apache-maven-3.8.4-bin.zip"),
		},
		"src/main/java/org/springframework/samples/petclinic/PetClinicApplication.java": &fstest.MapFile{
			Data: []byte(`
package org.springframework.samples.petclinic;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class PetClinicApplication {
    public static void main(String[] args) {
        SpringApplication.run(PetClinicApplication.class, args);
    }
}`),
		},
	}

	// Set up mock client
	mockClient := &llm.Mock{
		FactsJSON: `{
			"language": "java",
			"framework": "spring-boot",
			"build_tool": "maven",
			"build_cmd": "./mvnw -q package -DskipTests",
			"build_dir": ".",
			"start_cmd": "java -jar target/*.jar",
			"artifact": "target/*.jar",
			"ports": [8080],
			"health": "/actuator/health",
			"env": {},
			"base_image": "openjdk:11-jdk"
		}`,
		Dockerfile: `FROM openjdk:11-jdk
WORKDIR /app
COPY mvnw .
COPY .mvn .mvn
COPY pom.xml .
RUN chmod +x ./mvnw && ./mvnw -q package -DskipTests
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/actuator/health || exit 1
CMD ["java", "-jar", "target/*.jar"]`,
	}

	// Run the generation loop
	dockerfile, err := Run(context.Background(), fsys, mockClient)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify Dockerfile contents
	requiredCommands := []string{
		"FROM",
		"WORKDIR",
		"COPY",
		"RUN",
		"EXPOSE",
		"HEALTHCHECK",
		"CMD",
	}

	for _, cmd := range requiredCommands {
		if !strings.Contains(dockerfile, cmd) {
			t.Errorf("Dockerfile missing command: %s", cmd)
		}
	}

	// Verify Spring Boot specific elements
	springBootElements := []string{
		"openjdk:11-jdk",
		"./mvnw -q package -DskipTests",
		"target/*.jar",
		"8080",
		"/actuator/health",
	}

	for _, element := range springBootElements {
		if !strings.Contains(dockerfile, element) {
			t.Errorf("Dockerfile missing Spring Boot element: %s", element)
		}
	}
}
