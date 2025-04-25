package facts

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/aliou/dockerfile-gen/internal/detect"
)

func TestInfer(t *testing.T) {
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

	// Create test rule
	rule := detect.Rule{
		Name: "spring-boot",
		Tool: "maven",
	}

	// Test fact inference
	facts, err := Infer(context.Background(), fsys, rule)
	if err != nil {
		t.Fatalf("Infer() error = %v", err)
	}

	// Verify expected facts
	if facts.Language != "java" {
		t.Errorf("expected language java, got %s", facts.Language)
	}
	if facts.Framework != "spring-boot" {
		t.Errorf("expected framework spring-boot, got %s", facts.Framework)
	}
	if facts.BuildTool != "maven" {
		t.Errorf("expected build tool maven, got %s", facts.BuildTool)
	}
	if facts.BuildCmd != "./mvnw -q package -DskipTests" {
		t.Errorf("unexpected build command: %s", facts.BuildCmd)
	}
	if facts.StartCmd != "java -jar target/*.jar" {
		t.Errorf("unexpected start command: %s", facts.StartCmd)
	}
	if facts.Health != "/actuator/health" {
		t.Errorf("unexpected health endpoint: %s", facts.Health)
	}
}
