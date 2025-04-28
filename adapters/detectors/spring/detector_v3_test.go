package spring

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
)

func TestSpringBootDetectorV3_Detect(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		wantInfo  core.StackInfo
		wantFound bool
		wantErr   bool
	}{
		{
			name: "Maven project with all signals",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
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
    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>`,
				"src/main/java/com/example/Application.java": `package com.example;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}`,
				"src/main/resources/application.properties": "spring.application.name=test",
			},
			wantInfo: core.StackInfo{
				Name:       "spring-boot",
				BuildTool:  "maven",
				Port:       8080,
				Version:    "3.2.0",
				Confidence: 1.0,
				DetectedFiles: []string{
					"pom.xml",
				},
			},
			wantFound: true,
		},
		{
			name: "Gradle project with all signals",
			files: map[string]string{
				"build.gradle": `plugins {
    id 'org.springframework.boot' version '3.2.0'
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
}`,
				"src/main/java/com/example/Application.java": `package com.example;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}`,
				"src/main/resources/application.yml": `server:
  port: 9090`,
			},
			wantInfo: core.StackInfo{
				Name:       "spring-boot",
				BuildTool:  "gradle",
				Port:       8080,
				Version:    "3.2.0",
				Confidence: 1.0,
				DetectedFiles: []string{
					"build.gradle",
				},
			},
			wantFound: true,
		},
		{
			name: "Maven project with two signals",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
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
</project>`,
			},
			wantInfo: core.StackInfo{
				Name:       "spring-boot",
				BuildTool:  "maven",
				Port:       8080,
				Version:    "3.2.0",
				Confidence: 0.8,
				DetectedFiles: []string{
					"pom.xml",
				},
			},
			wantFound: true,
		},
		{
			name: "Gradle project with one signal",
			files: map[string]string{
				"build.gradle": `plugins {
    id 'org.springframework.boot' version '3.2.0'
}`,
			},
			wantInfo: core.StackInfo{
				Name:       "spring-boot",
				BuildTool:  "gradle",
				Port:       8080,
				Version:    "3.2.0",
				Confidence: 0.5,
				DetectedFiles: []string{
					"build.gradle",
				},
			},
			wantFound: true,
		},
		{
			name: "Mixed builders - prefer Maven",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
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
</project>`,
				"build.gradle": `plugins {
    id 'java'
}`,
			},
			wantInfo: core.StackInfo{
				Name:       "spring-boot",
				BuildTool:  "maven",
				Port:       8080,
				Version:    "3.2.0",
				Confidence: 0.8,
				DetectedFiles: []string{
					"pom.xml",
				},
			},
			wantFound: true,
		},
		{
			name: "Not a Spring Boot project",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.apache.maven</groupId>
        <artifactId>maven-parent</artifactId>
        <version>1.0</version>
    </parent>
</project>`,
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test filesystem
			fsys := fstest.MapFS{}
			for path, content := range tt.files {
				fsys[path] = &fstest.MapFile{
					Data: []byte(content),
				}
			}

			// Create detector and run test
			detector := NewSpringBootDetectorV3()
			info, found, err := detector.Detect(context.Background(), fsys, nil)

			// Check results
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.wantInfo, info)
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

func BenchmarkDetectSpringBoot(b *testing.B) {
	// Create test filesystems for different Spring Boot projects
	fsystems := []fstest.MapFS{
		// Simple Maven project
		{
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
		// Simple Gradle project
		{
			"build.gradle": &fstest.MapFile{
				Data: []byte(`plugins {
    id 'org.springframework.boot' version '3.2.0'
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
}`),
			},
		},
		// Deep nested Maven project
		{
			"services/api/pom.xml": &fstest.MapFile{
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
		// Deep nested Gradle project
		{
			"apps/payment/build.gradle": &fstest.MapFile{
				Data: []byte(`plugins {
    id 'org.springframework.boot' version '3.2.0'
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
}`),
			},
		},
		// Mixed builders project
		{
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
	}

	detector := NewSpringBootDetectorV3()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, fsys := range fsystems {
			detector.Detect(ctx, fsys, nil)
		}
	}
}
