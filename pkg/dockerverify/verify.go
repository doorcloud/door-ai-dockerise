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
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerClient is an interface for Docker operations
type DockerClient interface {
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

// Verify builds the Dockerfile in a temporary directory and returns the result
func Verify(ctx context.Context, fsys fs.FS, dockerfile string, timeout time.Duration) (bool, string, error) {
	// Create a temporary directory for the build
	dir, err := os.MkdirTemp("", "dockergen-*")
	if err != nil {
		return false, "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return false, "", fmt.Errorf("write Dockerfile: %w", err)
	}

	// Create .dockerignore
	if err := createDockerignore(dir); err != nil {
		return false, "", fmt.Errorf("create .dockerignore: %w", err)
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
		return false, "", fmt.Errorf("copy files: %w", err)
	}

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"),
	)
	if err != nil {
		return false, "", fmt.Errorf("create Docker client: %w", err)
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
		return false, "", fmt.Errorf("build image: %w", err)
	}
	defer resp.Body.Close()

	// Read build output
	var logs strings.Builder
	if _, err := io.Copy(&logs, resp.Body); err != nil {
		return false, "", fmt.Errorf("read build output: %w", err)
	}

	// Get the last 100 lines of logs
	logLines := strings.Split(logs.String(), "\n")
	start := len(logLines) - 100
	if start < 0 {
		start = 0
	}
	lastLogs := strings.Join(logLines[start:], "\n")

	// Check if build was successful
	if strings.Contains(logs.String(), "Successfully built") {
		return true, lastLogs, nil
	}

	return false, lastLogs, nil
}

// VerifyWithClient builds the Dockerfile in a temporary directory using the provided Docker client
func VerifyWithClient(ctx context.Context, fsys fs.FS, dockerfile string, timeout time.Duration, cli DockerClient) (bool, string, error) {
	// Create a temporary directory for the build
	dir, err := os.MkdirTemp("", "dockergen-*")
	if err != nil {
		return false, "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return false, "", fmt.Errorf("write Dockerfile: %w", err)
	}

	// Create .dockerignore
	if err := createDockerignore(dir); err != nil {
		return false, "", fmt.Errorf("create .dockerignore: %w", err)
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
		return false, "", fmt.Errorf("copy files: %w", err)
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
		return false, "", fmt.Errorf("build image: %w", err)
	}
	defer resp.Body.Close()

	// Read build output
	var logs strings.Builder
	if _, err := io.Copy(&logs, resp.Body); err != nil {
		return false, "", fmt.Errorf("read build output: %w", err)
	}

	// Get the last 100 lines of logs
	logLines := strings.Split(logs.String(), "\n")
	start := len(logLines) - 100
	if start < 0 {
		start = 0
	}
	lastLogs := strings.Join(logLines[start:], "\n")

	// Check if build was successful
	if strings.Contains(logs.String(), "Successfully built") {
		return true, lastLogs, nil
	}

	return false, lastLogs, nil
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
