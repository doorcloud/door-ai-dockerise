package detect

import (
	"io/fs"
)

func walk(fsys fs.FS, visit fs.WalkDirFunc) error {
	return fs.WalkDir(fsys, ".", visit)
}
