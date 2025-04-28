package spring

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestExtractor(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected Spec
	}{
		{
			name: "Maven with toolchains",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <properties>
        <java.version>21</java.version>
    </properties>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.6</version>
    </parent>
</project>`,
				".mvn/toolchains.xml": `<?xml version="1.0" encoding="UTF-8"?>
<toolchains>
    <toolchain>
        <type>jdk</type>
        <provides>
            <version>21</version>
        </provides>
    </toolchain>
</toolchains>`,
			},
			expected: Spec{
				JavaVersion:       "21",
				SpringBootVersion: strPtr("3.2.6"),
			},
		},
		{
			name: "Gradle Kotlin DSL",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    id("org.springframework.boot") version "3.2.6"
}
java {
    toolchain {
        languageVersion.set(JavaLanguageVersion.of(21))
    }
}`,
			},
			expected: Spec{
				JavaVersion:       "21",
				SpringBootVersion: strPtr("3.2.6"),
			},
		},
		{
			name: "Maven with dependencyManagement",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <properties>
        <java.version>17</java.version>
    </properties>
    <dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-dependencies</artifactId>
                <version>3.2.6</version>
                <type>pom</type>
                <scope>import</scope>
            </dependency>
        </dependencies>
    </dependencyManagement>
</project>`,
			},
			expected: Spec{
				JavaVersion:       "17",
				SpringBootVersion: strPtr("3.2.6"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			for name, content := range tt.files {
				fsys[name] = &fstest.MapFile{
					Data: []byte(content),
				}
			}

			extractor := NewExtractor()
			spec, err := extractor.Extract(fsys)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, *spec)
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestExtract_GradleMultiKtsToolchain(t *testing.T) {
	// Create test data directory
	projectPath := filepath.Join(os.TempDir(), "gradle_multi_kts_toolchain")
	err := os.MkdirAll(projectPath, 0o755)
	assert.NoError(t, err)
	defer os.RemoveAll(projectPath)

	// Create gradlew file
	err = os.WriteFile(filepath.Join(projectPath, "gradlew"), []byte("#!/bin/sh\necho 'Gradle Wrapper'"), 0o755)
	assert.NoError(t, err)

	// Create build.gradle.kts file
	err = os.WriteFile(filepath.Join(projectPath, "build.gradle.kts"), []byte(`
plugins {
	id("org.springframework.boot") version "3.2.0"
}

java {
	toolchain {
		languageVersion.set(JavaLanguageVersion.of(17))
	}
}`), 0o644)
	assert.NoError(t, err)

	extractor := NewExtractor()
	spec, err := extractor.Extract(projectPath)
	assert.NoError(t, err)
	assert.Equal(t, "gradle", spec.BuildTool)
	assert.Equal(t, "17", spec.JDKVersion)
	assert.Equal(t, "3.2.0", spec.SpringBootVersion)
	assert.Equal(t, "./gradlew build -x test", spec.BuildCmd)
	assert.Equal(t, "build/libs/*.jar", spec.Artifact)
	assert.Equal(t, "/actuator/health", spec.HealthEndpoint)
	assert.Equal(t, []int{8080}, spec.Ports)
}

func TestExtract_MavenParentDepMgmt(t *testing.T) {
	// Create test data directory
	projectPath := filepath.Join(os.TempDir(), "maven_parent_depMgmt")
	err := os.MkdirAll(projectPath, 0o755)
	assert.NoError(t, err)
	defer os.RemoveAll(projectPath)

	// Create pom.xml file
	err = os.WriteFile(filepath.Join(projectPath, "pom.xml"), []byte(`
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <name>demo</name>
    <description>Demo project for Spring Boot</description>
    <properties>
        <java.version>17</java.version>
    </properties>
</project>`), 0o644)
	assert.NoError(t, err)

	// Create target directory and SBOM file
	targetPath := filepath.Join(projectPath, "target")
	err = os.MkdirAll(targetPath, 0o755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(targetPath, "bom.cdx.json"), []byte("{}"), 0o644)
	assert.NoError(t, err)

	extractor := NewExtractor()
	spec, err := extractor.Extract(projectPath)
	assert.NoError(t, err)
	assert.Equal(t, "maven", spec.BuildTool)
	assert.Equal(t, "17", spec.JDKVersion)
	assert.Equal(t, "3.2.0", spec.SpringBootVersion)
	assert.Equal(t, "mvn clean package -DskipTests", spec.BuildCmd)
	assert.Equal(t, "target/*.jar", spec.Artifact)
	assert.Equal(t, "/actuator/health", spec.HealthEndpoint)
	assert.Equal(t, []int{8080}, spec.Ports)
	assert.Equal(t, filepath.Join(targetPath, "bom.cdx.json"), spec.Metadata["sbom_path"])
}

func TestExtract_SpringWithoutActuator(t *testing.T) {
	// Create test data directory
	projectPath := filepath.Join(os.TempDir(), "spring_without_actuator")
	err := os.MkdirAll(projectPath, 0o755)
	assert.NoError(t, err)
	defer os.RemoveAll(projectPath)

	// Create pom.xml file
	err = os.WriteFile(filepath.Join(projectPath, "pom.xml"), []byte(`
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <name>demo</name>
    <description>Demo project for Spring Boot</description>
    <properties>
        <java.version>17</java.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`), 0o644)
	assert.NoError(t, err)

	extractor := NewExtractor()
	spec, err := extractor.Extract(projectPath)
	assert.NoError(t, err)
	assert.Equal(t, "maven", spec.BuildTool)
	assert.Equal(t, "17", spec.JDKVersion)
	assert.Equal(t, "3.2.0", spec.SpringBootVersion)
	assert.Equal(t, "mvn clean package -DskipTests", spec.BuildCmd)
	assert.Equal(t, "target/*.jar", spec.Artifact)
	assert.Equal(t, "", spec.HealthEndpoint)
	assert.Equal(t, []int{8080}, spec.Ports)
}

func TestExtract_InvalidWarPackaging(t *testing.T) {
	// Create test data directory
	projectPath := filepath.Join(os.TempDir(), "invalid_war_packaging")
	err := os.MkdirAll(projectPath, 0o755)
	assert.NoError(t, err)
	defer os.RemoveAll(projectPath)

	// Create pom.xml file with WAR packaging
	err = os.WriteFile(filepath.Join(projectPath, "pom.xml"), []byte(`
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <packaging>war</packaging>
    <name>demo</name>
    <description>Demo project for Spring Boot</description>
    <properties>
        <java.version>17</java.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`), 0o644)
	assert.NoError(t, err)

	extractor := NewExtractor()
	spec, err := extractor.Extract(projectPath)
	assert.NoError(t, err)
	assert.Equal(t, "maven", spec.BuildTool)
	assert.Equal(t, "17", spec.JDKVersion)
	assert.Equal(t, "3.2.0", spec.SpringBootVersion)
	assert.Equal(t, "mvn clean package -DskipTests", spec.BuildCmd)
	assert.Equal(t, "target/*.war", spec.Artifact)
	assert.Equal(t, "/actuator/health", spec.HealthEndpoint)
	assert.Equal(t, []int{8080}, spec.Ports)
}
