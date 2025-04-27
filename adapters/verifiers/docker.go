package verifiers

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/client"
)

type Options struct {
	Socket  string
	LogSink io.Writer
	Timeout time.Duration
}

type Docker struct {
	client  *client.Client
	opts    Options
	logSink io.Writer
}

func NewDocker(opts Options) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.WithHost(opts.Socket))
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Docker{
		client:  cli,
		opts:    opts,
		logSink: opts.LogSink,
	}, nil
}

func (d *Docker) stream(line []byte) {
	if d.logSink != nil {
		d.logSink.Write(line)
	}
}

func (d *Docker) Verify(ctx context.Context, repoPath string, dockerfile string) error {
	// Build the Docker image with a tag
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", "test-app:latest", "-f", filepath.Join(repoPath, dockerfile), repoPath)

	// Stream build output
	if d.logSink != nil {
		buildPipe, err := buildCmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create build pipe: %w", err)
		}
		buildErrPipe, err := buildCmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create build error pipe: %w", err)
		}
		go func() {
			scanner := bufio.NewScanner(io.MultiReader(buildPipe, buildErrPipe))
			for scanner.Scan() {
				fmt.Fprintf(d.logSink, "docker build │ %s\n", scanner.Text())
			}
		}()
	} else {
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
	}

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// Run the container in detached mode
	runCmd := exec.CommandContext(ctx, "docker", "run", "-d", "--rm", "-p", "3000:3000", "test-app:latest")

	// Stream run output
	if d.logSink != nil {
		runPipe, err := runCmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create run pipe: %w", err)
		}
		runErrPipe, err := runCmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create run error pipe: %w", err)
		}
		go func() {
			scanner := bufio.NewScanner(io.MultiReader(runPipe, runErrPipe))
			for scanner.Scan() {
				fmt.Fprintf(d.logSink, "docker run   │ %s\n", scanner.Text())
			}
		}()
	} else {
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr
	}

	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("docker run failed: %w", err)
	}

	// Give the server some time to start
	time.Sleep(5 * time.Second)

	// Clean up the container
	cleanupCmd := exec.CommandContext(ctx, "docker", "ps", "-q", "--filter", "ancestor=test-app:latest")
	output, err := cleanupCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get container ID: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	if containerID != "" {
		stopCmd := exec.CommandContext(ctx, "docker", "stop", containerID)
		if err := stopCmd.Run(); err != nil {
			return fmt.Errorf("failed to stop container: %w", err)
		}
	}

	return nil
}

func createBuildContext(path string, dockerfile string) (io.ReadCloser, error) {
	// Create a temporary directory for the build context
	tmpDir, err := os.MkdirTemp("", "docker-build-")
	if err != nil {
		return nil, err
	}

	// Write Dockerfile
	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	// Copy the source directory
	if err := copyDir(path, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	// Create tarball
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(tmpDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
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
	}); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}
	if err := tw.Close(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	// Clean up temporary directory
	os.RemoveAll(tmpDir)

	return io.NopCloser(&buf), nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
