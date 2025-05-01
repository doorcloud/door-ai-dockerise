package spring

import (
	"context"
	"io/fs"
	"os"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
)

func TestSpringBootDetectorV3_Detect(t *testing.T) {
	tests := []struct {
		name   string
		fsys   fstest.MapFS
		want   core.StackInfo
		wantOk bool
	}{
		{
			name: "Maven project with all signals",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
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
</project>`),
				},
				"src/main/java/com/example/Application.java": &fstest.MapFile{
					Data: []byte(`package com.example;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}`),
				},
				"src/main/resources/application.properties": &fstest.MapFile{
					Data: []byte("spring.application.name=test"),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Version:       "3.2.0",
				Port:          8080,
				Confidence:    1.0,
				DetectedFiles: []string{"pom.xml"},
			},
			wantOk: true,
		},
		{
			name: "Gradle project with all signals",
			fsys: fstest.MapFS{
				"build.gradle": &fstest.MapFile{
					Data: []byte(`plugins {
    id 'org.springframework.boot' version '3.2.0'
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
}`),
				},
				"src/main/java/com/example/Application.java": &fstest.MapFile{
					Data: []byte(`package com.example;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}`),
				},
				"src/main/resources/application.yml": &fstest.MapFile{
					Data: []byte(`server:
  port: 9090`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"build.gradle"},
			},
			wantOk: true,
		},
		{
			name: "Maven project with two signals",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    0.8,
				DetectedFiles: []string{"pom.xml"},
			},
			wantOk: true,
		},
		{
			name: "Gradle project with one signal",
			fsys: fstest.MapFS{
				"build.gradle": &fstest.MapFile{
					Data: []byte(`plugins {
    id 'org.springframework.boot' version '3.2.0'
}`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    0.5,
				DetectedFiles: []string{"build.gradle"},
			},
			wantOk: true,
		},
		{
			name: "Mixed builders - prefer Maven",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`),
				},
				"build.gradle": &fstest.MapFile{
					Data: []byte(`plugins {
    id 'java'
}`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    0.8,
				DetectedFiles: []string{"pom.xml"},
			},
			wantOk: true,
		},
		{
			name: "Not a Spring Boot project",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.apache.maven</groupId>
        <artifactId>maven-parent</artifactId>
        <version>1.0</version>
    </parent>
</project>`),
				},
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpringBootDetectorV3()
			got, ok, err := detector.Detect(context.Background(), tt.fsys, nil)
			if err != nil {
				t.Errorf("SpringBootDetectorV3.Detect() error = %v", err)
				return
			}
			if ok != tt.wantOk {
				t.Errorf("SpringBootDetectorV3.Detect() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SpringBootDetectorV3.Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpringBootDetectorV3_ExtractVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		isMaven bool
	}{
		{
			name: "Maven version with snapshot",
			content: `<parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>3.2.0-SNAPSHOT</version>
</parent>`,
			want:    "3.2.0",
			isMaven: true,
		},
		{
			name: "Gradle version with snapshot",
			content: `plugins {
    id 'org.springframework.boot' version '3.2.0-SNAPSHOT'
}`,
			want:    "3.2.0",
			isMaven: false,
		},
		{
			name: "Maven version without snapshot",
			content: `<parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>3.2.0</version>
</parent>`,
			want:    "3.2.0",
			isMaven: true,
		},
		{
			name: "Gradle version without snapshot",
			content: `plugins {
    id 'org.springframework.boot' version '3.2.0'
}`,
			want:    "3.2.0",
			isMaven: false,
		},
		{
			name:    "No version found",
			content: "no version here",
			want:    "",
			isMaven: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpringBootDetectorV3()
			var got string
			if tt.isMaven {
				got = detector.extractVersionFromMaven(tt.content)
			} else {
				got = detector.extractVersionFromGradle(tt.content)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkDetect(b *testing.B) {
	detector := NewSpringBootDetectorV3()
	fsystems := []fs.FS{
		os.DirFS("testdata/spring/maven-single"),
		os.DirFS("testdata/spring/gradle-groovy"),
		os.DirFS("testdata/spring/gradle-kotlin"),
		os.DirFS("testdata/spring/maven-multi"),
		os.DirFS("testdata/spring/gradle-multi"),
		os.DirFS("testdata/spring_version_catalog_alias"),
		os.DirFS("testdata/settings_alias_plugin"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, fsys := range fsystems {
			detector.Detect(context.Background(), fsys, nil)
		}
	}
}

func TestIsSpringBoot(t *testing.T) {
	tests := []struct {
		name string
		path string
		fsys fstest.MapFS
		want bool
	}{
		{
			name: "maven single module",
			path: "pom.xml",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`),
				},
			},
			want: true,
		},
		{
			name: "Gradle project with all signals",
			path: "build.gradle",
			fsys: fstest.MapFS{
				"build.gradle": &fstest.MapFile{
					Data: []byte(`plugins {
    id 'org.springframework.boot' version '3.2.0'
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
}`),
				},
			},
			want: true,
		},
		{
			name: "Maven project with two signals",
			path: "pom.xml",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`),
				},
			},
			want: true,
		},
		{
			name: "Gradle project with one signal",
			path: "build.gradle",
			fsys: fstest.MapFS{
				"build.gradle": &fstest.MapFile{
					Data: []byte(`plugins {
    id 'org.springframework.boot' version '3.2.0'
}`),
				},
			},
			want: true,
		},
		{
			name: "Mixed builders - prefer Maven",
			path: "pom.xml",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`),
				},
				"build.gradle": &fstest.MapFile{
					Data: []byte(`plugins {
    id 'java'
}`),
				},
			},
			want: true,
		},
		{
			name: "Not a Spring Boot project",
			path: "pom.xml",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.apache.maven</groupId>
        <artifactId>maven-parent</artifactId>
        <version>1.0</version>
    </parent>
</project>`),
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpringBootDetectorV3()
			_, got, err := detector.Detect(context.Background(), tt.fsys, nil)
			if err != nil {
				t.Errorf("IsSpringBoot() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("IsSpringBoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectSpringBootRepos(t *testing.T) {
	tests := []struct {
		name string
		fsys fstest.MapFS
		want bool
	}{
		{
			name: "Spring Boot repository",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`),
				},
			},
			want: true,
		},
		{
			name: "Non-Spring Boot repository",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.apache.maven</groupId>
        <artifactId>maven-parent</artifactId>
        <version>1.0</version>
    </parent>
</project>`),
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpringBootDetectorV3()
			_, got, err := detector.Detect(context.Background(), tt.fsys, nil)
			if err != nil {
				t.Errorf("DetectSpringBootRepos() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("DetectSpringBootRepos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectSpringBootRepos_Integration(t *testing.T) {
	tests := []struct {
		name string
		fsys fstest.MapFS
		want bool
	}{
		{
			name: "Spring Boot repository",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`),
				},
			},
			want: true,
		},
		{
			name: "Non-Spring Boot repository",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.apache.maven</groupId>
        <artifactId>maven-parent</artifactId>
        <version>1.0</version>
    </parent>
</project>`),
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpringBootDetectorV3()
			_, got, err := detector.Detect(context.Background(), tt.fsys, nil)
			if err != nil {
				t.Errorf("DetectSpringBootRepos_Integration() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("DetectSpringBootRepos_Integration() = %v, want %v", got, tt.want)
			}
		})
	}
}
