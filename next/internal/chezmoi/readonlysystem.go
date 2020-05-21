package chezmoi

import (
	"os"
	"os/exec"
)

// A ReadOnlySystem is a system that may only be read from.
type ReadOnlySystem struct {
	s System
}

// NewReadOnlySystem returns a new ReadOnlySystem that wraps system.
func NewReadOnlySystem(system System) *ReadOnlySystem {
	return &ReadOnlySystem{
		s: system,
	}
}

// Chmod implements System.Chmod.
func (s *ReadOnlySystem) Chmod(name string, perm os.FileMode) error {
	return os.ErrPermission
}

// Delete implements System.Delete.
func (s *ReadOnlySystem) Delete(bucket, key []byte) error {
	return os.ErrPermission
}

// Get implements System.Get.
func (s *ReadOnlySystem) Get(bucket, key []byte) ([]byte, error) {
	return s.s.Get(bucket, key)
}

// Glob implements System.Glob.
func (s *ReadOnlySystem) Glob(pattern string) ([]string, error) {
	return s.s.Glob(pattern)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *ReadOnlySystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.s.IdempotentCmdOutput(cmd)
}

// Lstat implements System.Lstat.
func (s *ReadOnlySystem) Lstat(filename string) (os.FileInfo, error) {
	return s.s.Lstat(filename)
}

// Mkdir implements System.Mkdir.
func (s *ReadOnlySystem) Mkdir(name string, perm os.FileMode) error {
	return os.ErrPermission
}

// RawPath implements System.RawPath.
func (s *ReadOnlySystem) RawPath(path string) (string, error) {
	return s.s.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *ReadOnlySystem) ReadDir(dirname string) ([]os.FileInfo, error) {
	return s.s.ReadDir(dirname)
}

// ReadFile implements System.ReadFile.
func (s *ReadOnlySystem) ReadFile(filename string) ([]byte, error) {
	return s.s.ReadFile(filename)
}

// Readlink implements System.Readlink.
func (s *ReadOnlySystem) Readlink(name string) (string, error) {
	return s.s.Readlink(name)
}

// RemoveAll implements System.RemoveAll.
func (s *ReadOnlySystem) RemoveAll(name string) error {
	return os.ErrPermission
}

// Rename implements System.Rename.
func (s *ReadOnlySystem) Rename(oldpath, newpath string) error {
	return os.ErrPermission
}

// RunCmd implements System.RunCmd.
func (s *ReadOnlySystem) RunCmd(cmd *exec.Cmd) error {
	return os.ErrPermission
}

// RunScript implements System.RunScript.
func (s *ReadOnlySystem) RunScript(scriptname, dir string, data []byte) error {
	return os.ErrPermission
}

// Set implements System.Set.
func (s *ReadOnlySystem) Set(bucket, key, value []byte) error {
	return os.ErrPermission
}

// Stat implements System.Stat.
func (s *ReadOnlySystem) Stat(name string) (os.FileInfo, error) {
	return s.s.Stat(name)
}

// WriteFile implements System.WriteFile.
func (s *ReadOnlySystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.ErrPermission
}

// WriteSymlink implements System.WriteSymlink.
func (s *ReadOnlySystem) WriteSymlink(oldname, newname string) error {
	return os.ErrPermission
}
