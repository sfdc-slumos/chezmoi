package chezmoi

import (
	"path"
	"path/filepath"
)

// FIXME add PathHasPrefix

// PathJoin returns a clean, absolute path separated with forward
// slashes. If file is not an absolute path then it is joined on to dir.
func PathJoin(dir, file string) string {
	if !filepath.IsAbs(file) && !path.IsAbs(file) {
		file = filepath.Join(dir, file)
	}
	return filepath.ToSlash(filepath.Clean(file))
}

// PathsToSlashes returns a copy of paths with filepath.ToSlash to each element.
func PathsToSlashes(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, filepath.ToSlash(path))
	}
	return result
}
