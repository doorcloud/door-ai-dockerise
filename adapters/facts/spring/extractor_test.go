package spring

import (
	"io/fs"
	"os"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func stringPtr(s string) *string {
	return &s
}

func TestExtractor(t *testing.T) {
	tests := []struct {
		name     string
		fsys     fs.FS
		expected *Spec
	}{
		{
			name: "Maven with toolchains.xml",
			fsys: os.DirFS("testdata/maven_toolchains21"),
			expected: &Spec{
				BuildTool:         "maven",
				JavaVersion:       "21",
				SpringBootVersion: stringPtr("3.2.0"),
				BuildCmd:          "mvn clean package -DskipTests",
				Artifact:          "target/*.jar",
				HealthEndpoint:    "/actuator/health",
				Ports:             []int{8080},
			},
		},
		{
			name: "Gradle with Kotlin DSL toolchain",
			fsys: os.DirFS("testdata/gradle_kts_toolchain"),
			expected: &Spec{
				BuildTool:         "gradle",
				JavaVersion:       "21",
				SpringBootVersion: stringPtr("3.2.0"),
				BuildCmd:          "gradle build -x test",
				Artifact:          "build/libs/*.jar",
				HealthEndpoint:    "/actuator/health",
				Ports:             []int{8080},
			},
		},
		{
			name: "Maven with dependency management BOM",
			fsys: os.DirFS("testdata/depMgmt_bom"),
			expected: &Spec{
				BuildTool:         "maven",
				JavaVersion:       "17",
				SpringBootVersion: stringPtr("3.2.0"),
				BuildCmd:          "mvn clean package -DskipTests",
				Artifact:          "target/*.jar",
				HealthEndpoint:    "/actuator/health",
				Ports:             []int{8080},
			},
		},
		{
			name: "Spring Boot without actuator",
			fsys: os.DirFS("testdata/spring_without_actuator"),
			expected: &Spec{
				BuildTool:         "maven",
				JavaVersion:       "17",
				SpringBootVersion: stringPtr("3.2.0"),
				BuildCmd:          "mvn clean package -DskipTests",
				Artifact:          "target/*.jar",
				HealthEndpoint:    "",
				Ports:             []int{8080},
			},
		},
		{
			name: "Gradle multi-module with Kotlin DSL toolchain",
			fsys: os.DirFS("testdata/gradle_multi_kts_toolchain"),
			expected: &Spec{
				BuildTool:         "gradle",
				JavaVersion:       "21",
				SpringBootVersion: stringPtr("3.2.0"),
				BuildCmd:          "./gradlew api:build -x test",
				Artifact:          "api/build/libs/*.jar",
				HealthEndpoint:    "/actuator/health",
				Ports:             []int{8080},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExtractor()
			got, err := e.Extract(tt.fsys)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Extract() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtract_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create test filesystem
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`
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
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-actuator</artifactId>
        </dependency>
    </dependencies>
</project>`),
		},
		"target/bom.cdx.json": &fstest.MapFile{
			Data: []byte("{}"),
		},
	}

	extractor := NewExtractor()
	spec, err := extractor.Extract(fsys)
	assert.NoError(t, err)
	assert.Equal(t, "maven", spec.BuildTool)
	assert.Equal(t, "17", spec.JavaVersion)
	assert.Equal(t, "3.2.0", *spec.SpringBootVersion)
	assert.Equal(t, "mvn clean package -DskipTests", spec.BuildCmd)
	assert.Equal(t, "target/*.jar", spec.Artifact)
	assert.Equal(t, "/actuator/health", spec.HealthEndpoint)
	assert.Equal(t, []int{8080}, spec.Ports)
	assert.Equal(t, "target/bom.cdx.json", spec.Metadata["sbom_path"])
}
