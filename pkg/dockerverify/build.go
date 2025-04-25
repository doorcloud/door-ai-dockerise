package dockerverify

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
)

// Build builds a Docker image from the given Dockerfile and filesystem
func Build(ctx context.Context, fsys fs.FS, dockerfile string, cli APIClient) (string, error) {
	// Create a temporary directory for the build
	dir, err := os.MkdirTemp("", "dockergen-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return "", fmt.Errorf("write Dockerfile: %w", err)
	}

	// Create .dockerignore
	if err := os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte(`
.git
**/*.iml
.idea
docs
*.md
!mvnw
!.mvn/**
`), 0644); err != nil {
		return "", fmt.Errorf("create .dockerignore: %w", err)
	}

	// Copy files from the test filesystem to the temporary directory
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip Dockerfile and .dockerignore
		if path == "Dockerfile" || path == ".dockerignore" {
			return nil
		}

		// Create directories
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dir, path), 0755)
		}

		// Copy files
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		if err := os.WriteFile(filepath.Join(dir, path), data, 0644); err != nil {
			return fmt.Errorf("write file %s: %w", path, err)
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("copy files: %w", err)
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

	// Build the image
	resp, err := cli.ImageBuild(ctx, pr, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Remove:     true,
	})
	if err != nil {
		return "", fmt.Errorf("build image: %w", err)
	}
	defer resp.Body.Close()

	// Read build output
	var logs strings.Builder
	if _, err := io.Copy(&logs, resp.Body); err != nil {
		return "", fmt.Errorf("read build output: %w", err)
	}

	return logs.String(), nil
}
