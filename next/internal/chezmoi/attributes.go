package chezmoi

import (
	"strings"
)

// A SourceFileTargetType is a the type of a target represented by a file in the
// source state. A file in the source state can represent a file, script, or
// symlink in the target state.
type SourceFileTargetType int

// Source file types.
const (
	SourceFileTypeFile SourceFileTargetType = iota
	SourceFileTypeScript
	SourceFileTypeSymlink
)

// DirAttributes holds attributes parsed from a source directory name.
type DirAttributes struct {
	Name    string
	Exact   bool
	Private bool
}

// A FileAttributes holds attributes parsed from a source file name.
type FileAttributes struct {
	Name       string
	Type       SourceFileTargetType
	Empty      bool
	Encrypted  bool
	Executable bool
	Once       bool
	Private    bool
	Template   bool
}

// parseDirAttributes parses a single directory name in the source state.
func parseDirAttributes(sourceName string) DirAttributes {
	var (
		name    = sourceName
		exact   = false
		private = false
	)
	if strings.HasPrefix(name, exactPrefix) {
		name = strings.TrimPrefix(name, exactPrefix)
		exact = true
	}
	if strings.HasPrefix(name, privatePrefix) {
		name = strings.TrimPrefix(name, privatePrefix)
		private = true
	}
	if strings.HasPrefix(name, dotPrefix) {
		name = "." + strings.TrimPrefix(name, dotPrefix)
	}
	return DirAttributes{
		Name:    name,
		Exact:   exact,
		Private: private,
	}
}

// BaseName returns da's source name.
func (da DirAttributes) BaseName() string {
	sourceName := ""
	if da.Exact {
		sourceName += exactPrefix
	}
	if da.Private {
		sourceName += privatePrefix
	}
	if strings.HasPrefix(da.Name, ".") {
		sourceName += dotPrefix + strings.TrimPrefix(da.Name, ".")
	} else {
		sourceName += da.Name
	}
	return sourceName
}

// parseFileAttributes parses a source file name in the source state.
func parseFileAttributes(sourceName string) FileAttributes {
	var (
		name       = sourceName
		typ        = SourceFileTypeFile
		empty      = false
		encrypted  = false
		executable = false
		once       = false
		private    = false
		template   = false
	)
	switch {
	case strings.HasPrefix(name, runPrefix):
		name = strings.TrimPrefix(name, runPrefix)
		typ = SourceFileTypeScript
		if strings.HasPrefix(name, oncePrefix) {
			name = strings.TrimPrefix(name, oncePrefix)
			once = true
		}
	case strings.HasPrefix(name, symlinkPrefix):
		name = strings.TrimPrefix(name, symlinkPrefix)
		typ = SourceFileTypeSymlink
		if strings.HasPrefix(name, dotPrefix) {
			name = "." + strings.TrimPrefix(name, dotPrefix)
		}
	default:
		if strings.HasPrefix(name, encryptedPrefix) {
			name = strings.TrimPrefix(name, encryptedPrefix)
			encrypted = true
		}
		if strings.HasPrefix(name, privatePrefix) {
			name = strings.TrimPrefix(name, privatePrefix)
			private = true
		}
		if strings.HasPrefix(name, emptyPrefix) {
			name = strings.TrimPrefix(name, emptyPrefix)
			empty = true
		}
		if strings.HasPrefix(name, executablePrefix) {
			name = strings.TrimPrefix(name, executablePrefix)
			executable = true
		}
		if strings.HasPrefix(name, dotPrefix) {
			name = "." + strings.TrimPrefix(name, dotPrefix)
		}
	}
	if strings.HasSuffix(name, TemplateSuffix) {
		name = strings.TrimSuffix(name, TemplateSuffix)
		template = true
	}
	return FileAttributes{
		Name:       name,
		Type:       typ,
		Empty:      empty,
		Encrypted:  encrypted,
		Executable: executable,
		Once:       once,
		Private:    private,
		Template:   template,
	}
}

// BaseName returns fa's source name.
func (fa FileAttributes) BaseName() string {
	sourceName := ""
	switch fa.Type {
	case SourceFileTypeFile:
		if fa.Encrypted {
			sourceName += encryptedPrefix
		}
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.Empty {
			sourceName += emptyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeScript:
		sourceName = runPrefix
		if fa.Once {
			sourceName += oncePrefix
		}
	case SourceFileTypeSymlink:
		sourceName = symlinkPrefix
	}
	if strings.HasPrefix(fa.Name, ".") {
		sourceName += dotPrefix + strings.TrimPrefix(fa.Name, ".")
	} else {
		sourceName += fa.Name
	}
	if fa.Template {
		sourceName += TemplateSuffix
	}
	return sourceName
}
