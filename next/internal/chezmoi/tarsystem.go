package chezmoi

import (
	"archive/tar"
	"io"
	"os"
	"os/exec"
)

// A TARSystem is a System that writes to a TAR archive.
type TARSystem struct {
	nullReaderSystem
	w              *tar.Writer
	headerTemplate tar.Header
}

// NewTARSystem returns a new TARSystem that writes a TAR file to w.
func NewTARSystem(w io.Writer, headerTemplate tar.Header) *TARSystem {
	return &TARSystem{
		w:              tar.NewWriter(w),
		headerTemplate: headerTemplate,
	}
}

// Chmod implements System.Chmod.
func (s *TARSystem) Chmod(name string, mode os.FileMode) error {
	return os.ErrPermission
}

// Close closes m.
func (s *TARSystem) Close() error {
	return s.w.Close()
}

// Delete implements System.Delete.
func (s *TARSystem) Delete(bucket, key []byte) error {
	return os.ErrPermission
}

// Mkdir implements System.Mkdir.
func (s *TARSystem) Mkdir(name string, perm os.FileMode) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeDir
	header.Name = name + "/"
	header.Mode = int64(perm)
	return s.w.WriteHeader(&header)
}

// RemoveAll implements System.RemoveAll.
func (s *TARSystem) RemoveAll(name string) error {
	return os.ErrPermission
}

// Rename implements System.Rename.
func (s *TARSystem) Rename(oldpath, newpath string) error {
	return os.ErrPermission
}

// RunCmd implements System.RunCmd.
func (s *TARSystem) RunCmd(cmd *exec.Cmd) error {
	return nil
}

// RunScript implements System.RunScript.
func (s *TARSystem) RunScript(scriptname, dir string, data []byte) error {
	return s.WriteFile(scriptname, data, 0o700)
}

// Set implements System.Set.
func (s *TARSystem) Set(bucket, key, value []byte) error {
	return nil
}

// WriteFile implements System.WriteFile.
func (s *TARSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = filename
	header.Size = int64(len(data))
	header.Mode = int64(perm)
	if err := s.w.WriteHeader(&header); err != nil {
		return err
	}
	_, err := s.w.Write(data)
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *TARSystem) WriteSymlink(oldname, newname string) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeSymlink
	header.Name = newname
	header.Linkname = oldname
	return s.w.WriteHeader(&header)
}
