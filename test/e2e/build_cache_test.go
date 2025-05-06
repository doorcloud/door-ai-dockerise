//go:build integration
// +build integration

package e2e_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// copyDir copies a fixture so the second build runs in a fresh build context.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// requireDocker skips the test if Docker daemon isn't available.
func requireDocker(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker not running")
	}
}

func TestBuildCacheMounts(t *testing.T) {
	requireDocker(t)

	fixture := "testdata/spring_positive/petclinic_maven"
	tmp1 := t.TempDir()
	tmp2 := t.TempDir()
	_ = copyDir(fixture, tmp1)
	_ = copyDir(fixture, tmp2)

	run := func(dir string) (time.Duration, string) {
		cmd := exec.Command("go", "run", "./cmd/dockergen", dir)
		var buf bytes.Buffer
		cmd.Stdout, cmd.Stderr = &buf, &buf
		start := time.Now()
		if err := cmd.Run(); err != nil {
			t.Fatalf("build failed: %v\n%s", err, buf.String())
		}
		return time.Since(start), buf.String()
	}

	coldDur, _ := run(tmp1)    // first build (cold)
	warmDur, logs := run(tmp2) // second build should be fast

	if warmDur >= coldDur || warmDur > 90*time.Second {
		t.Fatalf("cache mount ineffective: cold=%v warm=%v", coldDur, warmDur)
	}

	if !(strings.Contains(logs, "--mount=type=cache,target=/root/.m2") ||
		strings.Contains(logs, "--mount=type=cache,target=/home/gradle/.gradle")) {
		t.Fatalf("cache mount string missing from logs")
	}
}
