package snippet

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// T represents a code snippet from a project file.
type T struct {
	// Path is the relative path to the file.
	Path string `json:"path"`

	// Content is the file content.
	Content string `json:"content"`

	// Language is the file's programming language.
	Language string `json:"language"`
}

// ReadFile reads a file and returns it as a snippet.
func ReadFile(path string) (T, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return T{}, fmt.Errorf("read file: %w", err)
	}

	ext := filepath.Ext(path)
	lang := strings.TrimPrefix(ext, ".")

	return T{
		Path:     path,
		Content:  string(content),
		Language: lang,
	}, nil
}

// ReadFileWithLimit reads a file and returns it as a snippet, limiting the content to maxLines.
func ReadFileWithLimit(path string, maxLines int) (T, error) {
	file, err := os.Open(path)
	if err != nil {
		return T{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for i := 0; i < maxLines && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return T{}, fmt.Errorf("scan file: %w", err)
	}

	ext := filepath.Ext(path)
	lang := strings.TrimPrefix(ext, ".")

	return T{
		Path:     path,
		Content:  strings.Join(lines, "\n"),
		Language: lang,
	}, nil
}

// ReadFiles reads multiple files and returns them as snippets.
func ReadFiles(paths []string) ([]T, error) {
	var snippets []T
	for _, path := range paths {
		snip, err := ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		snippets = append(snippets, snip)
	}
	return snippets, nil
}

// ReadFilesWithLimit reads multiple files and returns them as snippets, limiting each file to maxLines.
func ReadFilesWithLimit(paths []string, maxLines int) ([]T, error) {
	var snippets []T
	for _, path := range paths {
		snip, err := ReadFileWithLimit(path, maxLines)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		snippets = append(snippets, snip)
	}
	return snippets, nil
}

// FindFiles finds files matching the given patterns in the given directory.
func FindFiles(dir string, patterns []string) ([]string, error) {
	var paths []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", pattern, err)
		}
		paths = append(paths, matches...)
	}
	return paths, nil
}

// Log writes the snippets to the logger.
func Log(logger *slog.Logger, snippets []T) {
	if os.Getenv("DG_DEBUG") != "1" {
		return
	}

	for _, s := range snippets {
		logger.Debug("snippet",
			"path", s.Path,
			"language", s.Language,
			"content", truncate(s.Content, 200),
		)
	}
}

// truncate returns a truncated string with an ellipsis.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Write writes the snippets to the given writer.
func Write(w io.Writer, snippets []T) error {
	for _, s := range snippets {
		fmt.Fprintf(w, "=== %s ===\n", s.Path)
		fmt.Fprintln(w, s.Content)
		fmt.Fprintln(w)
	}
	return nil
}
