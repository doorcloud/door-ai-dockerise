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
</project>`), 0644)
			case "testdata/gradle":
				err = os.WriteFile(filepath.Join(projectPath, "build.gradle"), []byte(`
plugins {
	id 'org.springframework.boot' version '2.7.0'
}

java {
	sourceCompatibility = '11'
}`), 0644)
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
