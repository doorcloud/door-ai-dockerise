package dockerverify

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestCopyBuildContext_MavenWrapper(t *testing.T) {
	// Create test filesystem
	fsys := fstest.MapFS{
		"mvnw":               &fstest.MapFile{Data: []byte("#!/bin/sh")},
		"mvnw.cmd":           &fstest.MapFile{Data: []byte("@REM")},
		".mvn/wrapper/a.jar": &fstest.MapFile{Data: []byte("content")},
		"pom.xml":            &fstest.MapFile{Data: []byte("<project>")},
	}

	// Create temp dir
	dir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Copy build context
	if err := CopyBuildContext(fsys, dir); err != nil {
		t.Fatal(err)
	}

	// Create tar archive
	pr, pw := io.Pipe()
	go func() {
		tw := tar.NewWriter(pw)
		defer tw.Close()
		defer pw.Close()

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip root dir
			if path == dir {
				return nil
			}

			// Create header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			header.Name = relPath

			// Write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// Write file content
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				if _, err := io.Copy(tw, file); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	// Read tar and verify contents
	tr := tar.NewReader(pr)
	found := make(map[string]bool)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		found[header.Name] = true
	}

	// Check required files
	required := []string{
		"mvnw",
		"mvnw.cmd",
		".mvn/wrapper/a.jar",
	}
	for _, name := range required {
		if !found[name] {
			t.Errorf("file %s not found in tar", name)
		}
	}
}
