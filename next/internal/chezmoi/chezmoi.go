package chezmoi

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Configuration constants.
const (
	POSIXFileModes   = runtime.GOOS != "windows"
	Symlinks         = runtime.GOOS != "windows"
	PathSeparator    = '/'
	PathSeparatorStr = string(PathSeparator)
	ignorePrefix     = "."
)

// Configuration variables.
var (
	// DefaultTemplateOptions are the default template options.
	DefaultTemplateOptions = []string{"missingkey=error"}

	// DefaultUmask is the default umask.
	DefaultUmask = os.FileMode(0o22)

	scriptOnceStateBucket = []byte("script")
)

// Suffixes and prefixes.
const (
	dotPrefix        = "dot_"
	emptyPrefix      = "empty_"
	encryptedPrefix  = "encrypted_"
	exactPrefix      = "exact_"
	executablePrefix = "executable_"
	oncePrefix       = "once_"
	privatePrefix    = "private_"
	runPrefix        = "run_"
	symlinkPrefix    = "symlink_"
	TemplateSuffix   = ".tmpl"
)

// Special file names.
const (
	ChezmoiPrefix = ".chezmoi"

	dataName         = ChezmoiPrefix + "data"
	ignoreName       = ChezmoiPrefix + "ignore"
	removeName       = ChezmoiPrefix + "remove"
	templatesDirName = ChezmoiPrefix + "templates"
	versionName      = ChezmoiPrefix + "version"
)

var modeTypeNames = map[os.FileMode]string{
	0:                 "file",
	os.ModeDir:        "dir",
	os.ModeSymlink:    "symlink",
	os.ModeNamedPipe:  "named pipe",
	os.ModeSocket:     "socket",
	os.ModeDevice:     "device",
	os.ModeCharDevice: "char device",
}

type duplicateTargetError struct {
	targetName  string
	sourcePaths []string
}

func (e *duplicateTargetError) Error() string {
	return fmt.Sprintf("%s: duplicate target (%s)", e.targetName, strings.Join(e.sourcePaths, ", "))
}

type unsupportedFileTypeError struct {
	path string
	mode os.FileMode
}

func (e *unsupportedFileTypeError) Error() string {
	return fmt.Sprintf("%s: unsupported file type %s", e.path, modeTypeName(e.mode))
}

func modeTypeName(mode os.FileMode) string {
	if name, ok := modeTypeNames[mode&os.ModeType]; ok {
		return name
	}
	return fmt.Sprintf("0o%o: unknown type", mode&os.ModeType)
}
