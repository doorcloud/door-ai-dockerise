package fsutil

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Find returns paths matching any pattern up to maxDepth directories deep.
func Find(root string, maxDepth int, patterns ...string) ([]string, error) {
	var matches []string
	root = filepath.Clean(root)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate depth by counting path separators
		depth := strings.Count(path[len(root):], string(filepath.Separator))
		if depth > maxDepth {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				matches = append(matches, path)
				break
			}
		}

		return nil
	})

	return matches, err
}
