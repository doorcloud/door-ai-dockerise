package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_Spring(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test; set DG_E2E=1 to run")
	}

	// Create a temporary directory for the test
	dir := t.TempDir()

	// Create pom.xml
	err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>2.7.0</version>
        <relativePath/>
    </parent>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <name>demo</name>
    <description>Demo project for Spring Boot</description>
    <properties>
        <java.version>11</java.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-test</artifactId>
            <scope>test</scope>
        </dependency>
    </dependencies>
    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>`), 0644)
	assert.NoError(t, err)

	// Create src/main/java/com/example/demo/DemoApplication.java
	err = os.MkdirAll(filepath.Join(dir, "src", "main", "java", "com", "example", "demo"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "src", "main", "java", "com", "example", "demo", "DemoApplication.java"), []byte(`package com.example.demo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class DemoApplication {
    public static void main(String[] args) {
        SpringApplication.run(DemoApplication.class, args);
    }
}`), 0644)
	assert.NoError(t, err)

	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create pipeline
	p := pipeline.New(
		detectors.DefaultDetectors(),
		[]core.Generator{generate.NewLLM(mockLLM)},
		[]core.Verifier{verifiers.NewDocker()},
		[]core.FactProvider{facts.NewStatic()},
	)

	// Create log recorder
	reader, streamer := NewRecorder()

	// Run the pipeline
	ctx := context.Background()
	_, err = p.Run(ctx, dir, nil, streamer)
	assert.NoError(t, err)

	// Verify build logs
	AssertLog(t, reader, "mvn -q package", 2*time.Minute)
	AssertLog(t, reader, "BUILD SUCCESS", 2*time.Minute)
}
