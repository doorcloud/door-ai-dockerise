package facts

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	response string
}

func (c *mockClient) Chat(prompt string, model string) (string, error) {
	return c.response, nil
}

func TestInferWithClient(t *testing.T) {
	fsys := fstest.MapFS{
		"src/main/java/com/example/App.java": &fstest.MapFile{
			Data: []byte(`
package com.example;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class App {
    public static void main(String[] args) {
        SpringApplication.run(App.class, args);
    }
}
`),
		},
	}

	client := &mockClient{
		response: `{
			"language": "java",
			"framework": "spring-boot",
			"buildTool": "maven",
			"buildCmd": "./mvnw package",
			"buildDir": ".",
			"startCmd": "java -jar target/*.jar",
			"artifact": "target/*.jar",
			"ports": [8080],
			"health": "/actuator/health",
			"baseImage": "openjdk:11-jdk",
			"env": {}
		}`,
	}

	rule := detect.RuleInfo{
		Name: "spring-boot",
		Tool: "maven",
	}

	facts, err := InferWithClient(context.Background(), fsys, rule, client)
	assert.NoError(t, err)
	assert.Equal(t, "java", facts.Language)
	assert.Equal(t, "spring-boot", facts.Framework)
	assert.Equal(t, "maven", facts.BuildTool)
	assert.Equal(t, "./mvnw package", facts.BuildCmd)
	assert.Equal(t, ".", facts.BuildDir)
	assert.Equal(t, "java -jar target/*.jar", facts.StartCmd)
	assert.Equal(t, "target/*.jar", facts.Artifact)
	assert.Equal(t, []int{8080}, facts.Ports)
	assert.Equal(t, "/actuator/health", facts.Health)
	assert.Equal(t, "openjdk:11-jdk", facts.BaseImage)
	assert.Empty(t, facts.Env)
}

func TestGetFacts(t *testing.T) {
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
	xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>
	<parent>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-parent</artifactId>
		<version>2.7.0</version>
		<relativePath/> <!-- lookup parent from repository -->
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
	</dependencies>
</project>`),
		},
	}

	rule := detect.RuleInfo{
		Name: "spring-boot",
		Tool: "maven",
	}

	facts, err := GetFactsFromRule(fsys, rule)
	assert.NoError(t, err)
	assert.Equal(t, "java", facts.Language)
	assert.Equal(t, "spring-boot", facts.Framework)
	assert.Equal(t, "maven", facts.BuildTool)
	assert.Equal(t, "./mvnw -q package -DskipTests", facts.BuildCmd)
	assert.Equal(t, ".", facts.BuildDir)
	assert.Equal(t, "java -jar target/*.jar", facts.StartCmd)
	assert.Equal(t, "target/*.jar", facts.Artifact)
	assert.Equal(t, []int{8080}, facts.Ports)
	assert.Equal(t, "/actuator/health", facts.Health)
	assert.Equal(t, "openjdk:11-jdk", facts.BaseImage)
	assert.Empty(t, facts.Env)
}
