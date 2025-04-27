package detect

import (
	"io/fs"
)

// walk traverses a filesystem using fs.WalkDir
func walk(fsys fs.FS, visit fs.WalkDirFunc) error {
	return fs.WalkDir(fsys, ".", visit)
}
