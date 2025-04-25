package verify

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Verify builds the Dockerfile in a temporary directory and returns the result
func Verify(ctx context.Context, fsys fs.FS, dockerfile string) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Create a temporary directory for the build
	dir, err := os.MkdirTemp(os.TempDir(), "dockergen-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	// Set proper permissions on the temporary directory
	if err := os.Chmod(dir, 0777); err != nil {
		return fmt.Errorf("chmod temp dir: %w", err)
	}

	// First, create all directories with proper permissions
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			targetDir := filepath.Join(dir, path)
			if err := os.MkdirAll(targetDir, 0777); err != nil {
				return fmt.Errorf("create directory %s: %w", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	// Then copy all files
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip Dockerfile and .dockerignore
		if path == "Dockerfile" || path == ".dockerignore" {
			return nil
		}

		// Skip directories as they were already created
		if d.IsDir() {
			return nil
		}

		// Copy files
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		info, err := fs.Stat(fsys, path)
		if err != nil {
			return fmt.Errorf("stat file %s: %w", path, err)
		}

		targetFile := filepath.Join(dir, path)
		if err := os.WriteFile(targetFile, data, 0666); err != nil {
			return fmt.Errorf("write file %s: %w", path, err)
		}
		if err := os.Chmod(targetFile, info.Mode()|0666); err != nil {
			return fmt.Errorf("chmod file %s: %w", path, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("copy files: %w", err)
	}

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0666); err != nil {
		return fmt.Errorf("write Dockerfile: %w", err)
	}

	// Create .dockerignore
	if err := createDockerignore(dir); err != nil {
		return fmt.Errorf("create .dockerignore: %w", err)
	}

	// Ensure mvnw is executable if it exists
	if p := filepath.Join(dir, "mvnw"); fileExists(p) {
		if err := os.Chmod(p, 0755); err != nil {
			return fmt.Errorf("chmod mvnw: %w", err)
		}
	}

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"),
	)
	if err != nil {
		return fmt.Errorf("create Docker client: %w", err)
	}

	// Create a tar archive for the build context
	pr, pw := io.Pipe()
	go func() {
		tw := tar.NewWriter(pw)
		defer tw.Close()
		defer pw.Close()

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get the relative path
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			// Skip the root directory
			if relPath == "." {
				return nil
			}

			// Create a tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = relPath

			// Write the header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// If it's a file, write its contents
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

	// Build the image with timeout context
	resp, err := cli.ImageBuild(ctx, pr, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("build image: %w", err)
	}
	defer resp.Body.Close()

	// Read build output with timeout
	done := make(chan error)
	go func() {
		var logs strings.Builder
		_, err := io.Copy(&logs, resp.Body)
		if err != nil {
			done <- fmt.Errorf("read build output: %w", err)
			return
		}
		if !strings.Contains(logs.String(), "Successfully built") {
			done <- fmt.Errorf("build failed: %s", logs.String())
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("build timeout: %w", ctx.Err())
	}
}

// createDockerignore creates a .dockerignore file in the build context directory
func createDockerignore(dir string) error {
	ignoreContent := `.git
**/*.iml
.idea
docs
*.md
!mvnw
!.mvn/**
`
	return os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte(ignoreContent), 0644)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
