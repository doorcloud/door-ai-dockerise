package spring

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractor(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		wantSpec *Spec
		wantErr  bool
	}{
		{
			name:    "Maven project",
			project: "testdata/maven",
			wantSpec: &Spec{
				BuildTool:         "maven",
				JDKVersion:        "11",
				SpringBootVersion: "2.7.0",
				BuildCmd:          "mvn clean package -DskipTests",
				Artifact:          "target/*.jar",
				HealthEndpoint:    "/actuator/health",
				Ports:             []int{8080},
			},
			wantErr: false,
		},
		{
			name:    "Gradle project",
			project: "testdata/gradle",
			wantSpec: &Spec{
				BuildTool:         "gradle",
				JDKVersion:        "11",
				SpringBootVersion: "2.7.0",
				BuildCmd:          "./gradlew build -x test",
				Artifact:          "build/libs/*.jar",
				HealthEndpoint:    "/actuator/health",
				Ports:             []int{8080},
			},
			wantErr: false,
		},
		{
			name:     "Invalid project",
			project:  "testdata/invalid",
			wantSpec: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test data directory
			projectPath := filepath.Join(os.TempDir(), tt.name)
			err := os.MkdirAll(projectPath, 0755)
			assert.NoError(t, err)
			defer os.RemoveAll(projectPath)

			// Create build files based on test case
			switch tt.project {
			case "testdata/maven":
				err = os.WriteFile(filepath.Join(projectPath, "pom.xml"), []byte(`
<project>
	<properties>
		<java.version>11</java.version>
		<spring-boot.version>2.7.0</spring-boot.version>
	</properties>
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-actuator</artifactId>
		</dependency>
	</dependencies>
</project>`), 0644)
			case "testdata/gradle":
				err = os.WriteFile(filepath.Join(projectPath, "build.gradle"), []byte(`
plugins {
	id 'org.springframework.boot' version '2.7.0'
}

dependencies {
	implementation 'org.springframework.boot:spring-boot-starter-actuator'
}

java {
	sourceCompatibility = '11'
}`), 0644)
				// Create gradlew file
				err = os.WriteFile(filepath.Join(projectPath, "gradlew"), []byte("#!/bin/sh\necho 'Gradle Wrapper'"), 0755)
			}
			assert.NoError(t, err)

			// Run extractor
			extractor := NewExtractor()
			gotSpec, err := extractor.Extract(projectPath)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotSpec)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSpec, gotSpec)
			}
		})
	}
}

func TestExtract_GradleMultiKtsToolchain(t *testing.T) {
	// Create test data directory
	projectPath := filepath.Join(os.TempDir(), "gradle_multi_kts_toolchain")
	err := os.MkdirAll(projectPath, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(projectPath)

	// Create gradlew file
	err = os.WriteFile(filepath.Join(projectPath, "gradlew"), []byte("#!/bin/sh\necho 'Gradle Wrapper'"), 0755)
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
}`), 0644)
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
	err := os.MkdirAll(projectPath, 0755)
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
</project>`), 0644)
	assert.NoError(t, err)

	// Create target directory and SBOM file
	targetPath := filepath.Join(projectPath, "target")
	err = os.MkdirAll(targetPath, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(targetPath, "bom.cdx.json"), []byte("{}"), 0644)
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
	err := os.MkdirAll(projectPath, 0755)
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
</project>`), 0644)
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
	err := os.MkdirAll(projectPath, 0755)
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
</project>`), 0644)
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
