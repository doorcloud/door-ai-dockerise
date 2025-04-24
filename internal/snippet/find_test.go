package snippet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectBuildDir_MultiPom(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a root pom.xml without Spring Boot
	rootPom := filepath.Join(tmpDir, "pom.xml")
	if err := os.WriteFile(rootPom, []byte(`
<project>
    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
        </dependency>
    </dependencies>
</project>`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a module with Spring Boot
	moduleDir := filepath.Join(tmpDir, "module")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatal(err)
	}

	modulePom := filepath.Join(moduleDir, "pom.xml")
	if err := os.WriteFile(modulePom, []byte(`
<project>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	buildDir := DetectBuildDir(tmpDir)
	if buildDir != moduleDir {
		t.Errorf("DetectBuildDir() = %v, want %v", buildDir, moduleDir)
	}
}

// ... existing code ...
