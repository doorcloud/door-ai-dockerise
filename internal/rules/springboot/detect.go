package springboot

import (
	"io/fs"
	"os"
	"strings"
)

type Detector struct {
	HasMavenWrapper bool
}

func (d *Detector) Detect(fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() {
			return nil
		}

		// Check for Maven wrapper files
		if strings.HasSuffix(path, "mvnw") ||
			strings.HasSuffix(path, "mvnw.cmd") ||
			strings.Contains(path, string(os.PathSeparator)+".mvn"+string(os.PathSeparator)) {
			d.HasMavenWrapper = true
		}

		return nil
	})
}
