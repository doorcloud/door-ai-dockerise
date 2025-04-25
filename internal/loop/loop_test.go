package loop

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
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

	// Run the generation loop
	dockerfile, err := Run(context.Background(), fsys)
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
		"eclipse-temurin:17-jdk",
		"./mvnw -q package",
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
