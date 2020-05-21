// +build !windows

package chezmoi

import (
	"path"
)

// FIXME add PathHasPrefix

// PathJoin returns a clean, absolute path. If file is not an absolute path then
// it is joined on to dir.
func PathJoin(dir, file string) string {
	if !path.IsAbs(file) {
		file = path.Join(dir, file)
	}
	return path.Clean(file)
}

// PathsToSlashes returns paths.
func PathsToSlashes(paths []string) []string {
	return paths
}
