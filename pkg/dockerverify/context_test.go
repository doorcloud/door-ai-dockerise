package dockerverify

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestCopyBuildContext_withWrapper(t *testing.T) {
	// Create test filesystem with Maven wrapper
	fsys := fstest.MapFS{
		"mvnw": &fstest.MapFile{
			Data: []byte("#!/bin/sh\n# Maven wrapper script"),
			Mode: 0755,
		},
		".mvn/wrapper/maven-wrapper.jar": &fstest.MapFile{
			Data: []byte("mock jar file"),
		},
		".mvn/wrapper/maven-wrapper.properties": &fstest.MapFile{
			Data: []byte("distributionUrl=https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.8.4/apache-maven-3.8.4-bin.zip"),
		},
		"pom.xml": &fstest.MapFile{
			Data: []byte("<project></project>"),
		},
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "dockerverify-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy build context
	if err := copyBuildContext(fsys, tmpDir); err != nil {
		t.Fatalf("copyBuildContext failed: %v", err)
	}

	// Verify wrapper files were copied
	files := []string{
		"mvnw",
		".mvn/wrapper/maven-wrapper.jar",
		".mvn/wrapper/maven-wrapper.properties",
		"pom.xml",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s not found", file)
		}
	}

	// Verify mvnw is executable
	info, err := os.Stat(filepath.Join(tmpDir, "mvnw"))
	if err != nil {
		t.Fatalf("Failed to stat mvnw: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("mvnw is not executable")
	}
}
