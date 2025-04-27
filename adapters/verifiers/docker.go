package verifiers

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Options struct {
	Socket  string
	LogSink io.Writer // can be nil
	Timeout time.Duration
}

type Docker struct {
	client *client.Client
	opts   Options
}

func NewDocker(opts Options) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.WithHost(opts.Socket))
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Docker{
		client: cli,
		opts:   opts,
	}, nil
}

func (d *Docker) stream(line []byte) {
	if d.opts.LogSink != nil {
		d.opts.LogSink.Write(line)
	}
}

func (d *Docker) Verify(ctx context.Context, path string, dockerfile string) error {
	ctx, cancel := context.WithTimeout(ctx, d.opts.Timeout)
	defer cancel()

	buildCtx, err := createBuildContext(path, dockerfile)
	if err != nil {
		return err
	}
	defer buildCtx.Close()

	opts := types.ImageBuildOptions{
		Dockerfile:  "Dockerfile",
		Tags:        []string{"test-image"},
		Remove:      true,
		ForceRemove: true,
	}

	resp, err := d.client.ImageBuild(ctx, buildCtx, opts)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	// Stream build output
	decoder := json.NewDecoder(resp.Body)
	for {
		var msg struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode build output: %w", err)
		}
		if msg.Error != "" {
			return fmt.Errorf("build error: %s", msg.Error)
		}
		d.stream([]byte(msg.Stream))
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
