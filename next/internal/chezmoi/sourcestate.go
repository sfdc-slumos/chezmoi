package chezmoi

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/coreos/go-semver/semver"
	vfs "github.com/twpayne/go-vfs"
	"go.uber.org/multierr"
)

// A SourceState is a source state.
type SourceState struct {
	entries              map[string]SourceStateEntry
	system               System
	sourcePath           string
	umask                os.FileMode
	encryptionTool       EncryptionTool
	ignore               *PatternSet
	minVersion           semver.Version
	priorityTemplateData map[string]interface{}
	remove               *PatternSet
	templateData         map[string]interface{}
	templateFuncs        template.FuncMap
	templateOptions      []string
	templates            map[string]*template.Template
}

// A SourceStateOption sets an option on a source state.
type SourceStateOption func(*SourceState)

// WithEncryptionTool set the encryption tool.
func WithEncryptionTool(encryptionTool EncryptionTool) SourceStateOption {
	return func(s *SourceState) {
		s.encryptionTool = encryptionTool
	}
}

// WithPriorityTemplateData adds priority template data.
func WithPriorityTemplateData(priorityTemplateData map[string]interface{}) SourceStateOption {
	return func(s *SourceState) {
		recursiveMerge(s.priorityTemplateData, priorityTemplateData)
		recursiveMerge(s.templateData, s.priorityTemplateData)
	}
}

// WithSourcePath sets the source path.
func WithSourcePath(sourcePath string) SourceStateOption {
	return func(s *SourceState) {
		s.sourcePath = sourcePath
	}
}

// WithSystem sets the system.
func WithSystem(system System) SourceStateOption {
	return func(s *SourceState) {
		s.system = system
	}
}

// WithTemplateData adds template data.
func WithTemplateData(templateData map[string]interface{}) SourceStateOption {
	return func(s *SourceState) {
		recursiveMerge(s.templateData, templateData)
		recursiveMerge(s.templateData, s.priorityTemplateData)
	}
}

// WithTemplateFuncs sets the template functions.
func WithTemplateFuncs(templateFuncs template.FuncMap) SourceStateOption {
	return func(s *SourceState) {
		s.templateFuncs = templateFuncs
	}
}

// WithTemplateOptions sets the template options.
func WithTemplateOptions(templateOptions []string) SourceStateOption {
	return func(s *SourceState) {
		s.templateOptions = templateOptions
	}
}

// WithUmask sets the umask.
func WithUmask(umask os.FileMode) SourceStateOption {
	return func(s *SourceState) {
		s.umask = umask
	}
}

// NewSourceState creates a new source state with the given options.
func NewSourceState(options ...SourceStateOption) *SourceState {
	s := &SourceState{
		entries:              make(map[string]SourceStateEntry),
		umask:                DefaultUmask,
		encryptionTool:       &nullEncryptionTool{},
		ignore:               NewPatternSet(),
		priorityTemplateData: make(map[string]interface{}),
		remove:               NewPatternSet(),
		templateData:         make(map[string]interface{}),
		templateOptions:      DefaultTemplateOptions,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// AddOptions are options to SourceState.Add.
type AddOptions struct {
	AutoTemplate bool
	Empty        bool
	Encrypt      bool
	Exact        bool
	Include      *IncludeSet
	Template     bool
}

// Add adds sourceStateEntry to s.
func (s *SourceState) Add(sourceSystem System, destDir string, destPathInfos map[string]os.FileInfo, options *AddOptions) error {
	destPaths := make([]string, 0, len(destPathInfos))
	for destPath := range destPathInfos {
		destPaths = append(destPaths, destPath)
	}
	sort.Strings(destPaths)
	targetSourceState := &SourceState{
		entries: make(map[string]SourceStateEntry),
	}
	for _, destPath := range destPaths {
		// FIXME rename/remove old
		targetName := strings.TrimPrefix(destPath, destDir+PathSeparatorStr)
		sourceStateEntry, err := s.sourceStateEntry(sourceSystem, destPath, destPathInfos[destPath], options)
		if err != nil {
			return err
		}
		if sourceStateEntry != nil {
			targetSourceState.entries[targetName] = sourceStateEntry
		}
	}
	// FIXME include
	return targetSourceState.ApplyAll(sourceSystem, s.sourcePath, options.Include)
}

// ApplyAll updates targetDir in fs to match s.
func (s *SourceState) ApplyAll(targetSystem System, targetDir string, include *IncludeSet) error {
	for _, targetName := range s.sortedTargetNames() {
		if err := s.ApplyOne(targetSystem, targetDir, targetName, include); err != nil {
			return err
		}
	}
	return nil
}

// ApplyOne updates targetName in targetDir on fs to match s using s.
func (s *SourceState) ApplyOne(targetSystem System, targetDir, targetName string, include *IncludeSet) error {
	targetStateEntry, err := s.entries[targetName].TargetStateEntry()
	if err != nil {
		return err
	}

	if !include.IncludeTargetStateEntry(targetStateEntry) {
		return nil
	}

	targetPath := path.Join(targetDir, targetName)
	destStateEntry, err := NewDestStateEntry(targetSystem, targetPath)
	if err != nil {
		return err
	}

	if err := targetStateEntry.Apply(targetSystem, destStateEntry); err != nil {
		return err
	}

	if targetStateDir, ok := targetStateEntry.(*TargetStateDir); ok {
		if targetStateDir.exact {
			infos, err := targetSystem.ReadDir(targetPath)
			if err != nil {
				return err
			}
			baseNames := make([]string, 0, len(infos))
			for _, info := range infos {
				if baseName := info.Name(); baseName != "." && baseName != ".." {
					baseNames = append(baseNames, baseName)
				}
			}
			sort.Strings(baseNames)
			for _, baseName := range baseNames {
				if _, ok := s.entries[path.Join(targetName, baseName)]; !ok {
					if err := targetSystem.RemoveAll(path.Join(targetPath, baseName)); err != nil {
						return err
					}
				}
			}
		}
	}

	// FIXME chezmoiremove
	return nil
}

// Entries returns s's source state entries.
func (s *SourceState) Entries() map[string]SourceStateEntry {
	return s.entries
}

// Ignored returns if targetName is ignored.
func (s *SourceState) Ignored(targetName string) bool {
	return s.ignore.Match(targetName)
}

// TargetNames returns all of s's target names in alphabetical order.
func (s *SourceState) TargetNames() []string {
	targetNames := make([]string, 0, len(s.entries))
	for targetName := range s.entries {
		targetNames = append(targetNames, targetName)
	}
	sort.Strings(targetNames)
	return targetNames
}

// Entry returns the source state entry for targetName.
func (s *SourceState) Entry(targetName string) (SourceStateEntry, bool) {
	sourceStateEntry, ok := s.entries[targetName]
	return sourceStateEntry, ok
}

// Evaluate evaluates every target state entry in s.
func (s *SourceState) Evaluate() error {
	for _, targetName := range s.sortedTargetNames() {
		sourceStateEntry := s.entries[targetName]
		if err := sourceStateEntry.Evaluate(); err != nil {
			return err
		}
		targetStateEntry, err := sourceStateEntry.TargetStateEntry()
		if err != nil {
			return err
		}
		if err := targetStateEntry.Evaluate(); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTemplateData returns the result of executing template data.
func (s *SourceState) ExecuteTemplateData(name string, data []byte) ([]byte, error) {
	tmpl, err := template.New(name).Option(s.templateOptions...).Funcs(s.templateFuncs).Parse(string(data))
	if err != nil {
		return nil, err
	}
	for name, t := range s.templates {
		tmpl, err = tmpl.AddParseTree(name, t.Tree)
		if err != nil {
			return nil, err
		}
	}
	output := &bytes.Buffer{}
	if err = tmpl.ExecuteTemplate(output, name, s.TemplateData()); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// MinVersion returns the minimum version for which s is valid.
func (s *SourceState) MinVersion() semver.Version {
	return s.minVersion
}

// MustEntry returns the source state entry associated with targetName, and
// panics if it does not exist.
func (s *SourceState) MustEntry(targetName string) SourceStateEntry {
	sourceStateEntry, ok := s.entries[targetName]
	if !ok {
		panic(fmt.Sprintf("%s: no source state entry", targetName))
	}
	return sourceStateEntry
}

// Read reads a source state from sourcePath.
func (s *SourceState) Read() error {
	info, err := s.system.Lstat(s.sourcePath)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("%s: not a directory", s.sourcePath)
	}

	// Read all source entries.
	allSourceStateEntries := make(map[string][]SourceStateEntry)
	sourceDirPrefix := s.sourcePath + PathSeparatorStr
	if err := vfs.WalkSlash(s.system, s.sourcePath, func(sourcePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if sourcePath == s.sourcePath {
			return nil
		}
		relPath := strings.TrimPrefix(sourcePath, sourceDirPrefix)
		sourceDirName, sourceName := path.Split(relPath)
		targetDirName := getTargetDirName(sourceDirName)
		switch {
		case strings.HasPrefix(info.Name(), dataName):
			return s.addTemplateData(sourcePath)
		case info.Name() == ignoreName:
			// .chezmoiignore is interpreted as a template. vfs.WalkSlash walks
			// in alphabetical order, so, luckily for us, .chezmoidata will be
			// read before .chezmoiignore, so data in .chezmoidata is available
			// to be used in .chezmoiignore. Unluckily for us, .chezmoitemplates
			// will be read afterwards so partial templates will not be
			// available in .chezmoiignore.
			return s.addPatterns(s.ignore, sourcePath, sourceDirName)
		case info.Name() == removeName:
			// The comment about .chezmoiignore applies to .chezmoiremove too.
			return s.addPatterns(s.remove, sourcePath, targetDirName)
		case info.Name() == templatesDirName:
			if err := s.addTemplatesDir(sourcePath); err != nil {
				return err
			}
			return vfs.SkipDir
		case info.Name() == versionName:
			return s.addVersionFile(sourcePath)
		case strings.HasPrefix(info.Name(), ChezmoiPrefix):
			fallthrough
		case strings.HasPrefix(info.Name(), ignorePrefix):
			if info.IsDir() {
				return vfs.SkipDir
			}
			return nil
		case info.IsDir():
			da := parseDirAttributes(sourceName)
			targetName := path.Join(targetDirName, da.Name)
			if s.ignore.Match(targetName) {
				return nil
			}
			sourceStateEntry := s.newSourceStateDir(sourcePath, da)
			allSourceStateEntries[targetName] = append(allSourceStateEntries[targetName], sourceStateEntry)
			return nil
		case info.Mode().IsRegular():
			fa := parseFileAttributes(sourceName)
			targetName := path.Join(targetDirName, fa.Name)
			if s.ignore.Match(targetName) {
				return nil
			}
			sourceStateEntry := s.newSourceStateFile(sourcePath, fa, targetName)
			allSourceStateEntries[targetName] = append(allSourceStateEntries[targetName], sourceStateEntry)
			return nil
		default:
			return &unsupportedFileTypeError{
				path: sourcePath,
				mode: info.Mode(),
			}
		}
	}); err != nil {
		return err
	}

	// Checking for duplicate source entries with the same target name. Iterate
	// over the target names in order so that any error is deterministic.
	targetNames := make([]string, 0, len(allSourceStateEntries))
	for targetName := range allSourceStateEntries {
		targetNames = append(targetNames, targetName)
	}
	sort.Strings(targetNames)
	for _, targetName := range targetNames {
		sourceStateEntries := allSourceStateEntries[targetName]
		if len(sourceStateEntries) == 1 {
			continue
		}
		sourcePaths := make([]string, 0, len(sourceStateEntries))
		for _, sourceStateEntry := range sourceStateEntries {
			sourcePaths = append(sourcePaths, sourceStateEntry.Path())
		}
		err = multierr.Append(err, &duplicateTargetError{
			targetName:  targetName,
			sourcePaths: sourcePaths,
		})
	}
	if err != nil {
		return err
	}

	// Populate s.Entries with the unique source entry for each target.
	for targetName, sourceEntries := range allSourceStateEntries {
		s.entries[targetName] = sourceEntries[0]
	}

	return nil
}

// Remove removes everything in targetDir that matches s's remove pattern set.
func (s *SourceState) Remove(system System, targetDir string) error {
	// Build a set of targets to remove.
	targetDirPrefix := targetDir + PathSeparatorStr
	targetPathsToRemove := newStringSet()
	for include := range s.remove.includes {
		matches, err := system.Glob(path.Join(targetDir, include))
		if err != nil {
			return err
		}
		for _, match := range matches {
			// Don't remove targets that are excluded from remove.
			if !s.remove.Match(strings.TrimPrefix(match, targetDirPrefix)) {
				continue
			}
			targetPathsToRemove.Add(match)
		}
	}

	// Remove targets in order. Parent directories are removed before their
	// children, which is okay because RemoveAll does not treat os.ErrNotExist
	// as an error.
	sortedTargetPathsToRemove := targetPathsToRemove.Elements()
	sort.Strings(sortedTargetPathsToRemove)
	for _, targetPath := range sortedTargetPathsToRemove {
		if err := system.RemoveAll(targetPath); err != nil {
			return err
		}
	}
	return nil
}

// TemplateData returns s's template data.
func (s *SourceState) TemplateData() map[string]interface{} {
	return s.templateData
}

func (s *SourceState) addPatterns(patternSet *PatternSet, sourcePath, relPath string) error {
	data, err := s.executeTemplate(sourcePath)
	if err != nil {
		return err
	}
	dir := path.Dir(relPath)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		text := scanner.Text()
		if index := strings.IndexRune(text, '#'); index != -1 {
			text = text[:index]
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		include := true
		if strings.HasPrefix(text, "!") {
			include = false
			text = strings.TrimPrefix(text, "!")
		}
		pattern := path.Join(dir, text)
		if err := patternSet.Add(pattern, include); err != nil {
			return fmt.Errorf("%s: %w", sourcePath, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s: %w", sourcePath, err)
	}
	return nil
}

func (s *SourceState) addTemplateData(sourcePath string) error {
	_, name := path.Split(sourcePath)
	suffix := strings.TrimPrefix(name, dataName+".")
	format, ok := Formats[strings.ToLower(suffix)]
	if !ok {
		return fmt.Errorf("%s: unknown format", sourcePath)
	}
	data, err := s.system.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("%s: %w", sourcePath, err)
	}
	var templateData map[string]interface{}
	if err := format.Decode(data, &templateData); err != nil {
		return fmt.Errorf("%s: %w", sourcePath, err)
	}
	recursiveMerge(s.templateData, templateData)
	recursiveMerge(s.templateData, s.priorityTemplateData)
	return nil
}

func (s *SourceState) addTemplatesDir(templateDir string) error {
	templateDirPrefix := templateDir + PathSeparatorStr
	return vfs.WalkSlash(s.system, templateDir, func(templatePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch {
		case info.Mode().IsRegular():
			contents, err := s.system.ReadFile(templatePath)
			if err != nil {
				return err
			}
			name := strings.TrimPrefix(templatePath, templateDirPrefix)
			tmpl, err := template.New(name).Parse(string(contents))
			if err != nil {
				return err
			}
			if s.templates == nil {
				s.templates = make(map[string]*template.Template)
			}
			s.templates[name] = tmpl
			return nil
		case info.IsDir():
			return nil
		default:
			return &unsupportedFileTypeError{
				path: templatePath,
				mode: info.Mode(),
			}
		}
	})
}

// addVersionFile reads a .chezmoiversion file from source path and updates s's
// minimum version if it contains a more recent version than the current minimum
// version.
func (s *SourceState) addVersionFile(sourcePath string) error {
	data, err := s.system.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	version, err := semver.NewVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}
	if s.minVersion.LessThan(*version) {
		s.minVersion = *version
	}
	return nil
}

func (s *SourceState) executeTemplate(path string) ([]byte, error) {
	data, err := s.system.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return s.ExecuteTemplateData(path, data)
}

func (s *SourceState) newSourceStateDir(sourcePath string, da DirAttributes) *SourceStateDir {
	perm := os.FileMode(0o777)
	if da.Private {
		perm &^= 0o77
	}
	perm &^= s.umask

	targetStateDir := &TargetStateDir{
		perm:  perm,
		exact: da.Exact,
	}

	return &SourceStateDir{
		path:             sourcePath,
		Attributes:       da,
		targetStateEntry: targetStateDir,
	}
}

func (s *SourceState) newSourceStateFile(sourcePath string, fa FileAttributes, targetName string) *SourceStateFile {
	lazyContents := &lazyContents{
		contentsFunc: func() ([]byte, error) {
			contents, err := s.system.ReadFile(sourcePath)
			if err != nil {
				return nil, err
			}
			if !fa.Encrypted {
				return contents, nil
			}
			// FIXME pass targetName as filenameHint
			return s.encryptionTool.Decrypt(sourcePath, contents)
		},
	}

	var targetStateEntryFunc func() (TargetStateEntry, error)
	switch fa.Type {
	case SourceFileTypeFile:
		targetStateEntryFunc = func() (TargetStateEntry, error) {
			contents, err := lazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fa.Template {
				contents, err = s.ExecuteTemplateData(sourcePath, contents)
				if err != nil {
					return nil, err
				}
			}
			if !fa.Empty && isEmpty(contents) {
				return &TargetStateAbsent{}, nil
			}
			perm := os.FileMode(0o666)
			if fa.Executable {
				perm |= 0o111
			}
			if fa.Private {
				perm &^= 0o77
			}
			perm &^= s.umask
			return &TargetStateFile{
				lazyContents: newLazyContents(contents),
				perm:         perm,
			}, nil
		}
	case SourceFileTypeScript:
		targetStateEntryFunc = func() (TargetStateEntry, error) {
			contents, err := lazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fa.Template {
				contents, err = s.ExecuteTemplateData(sourcePath, contents)
				if err != nil {
					return nil, err
				}
			}
			return &TargetStateScript{
				lazyContents: newLazyContents(contents),
				name:         targetName,
				once:         fa.Once,
			}, nil
		}
	case SourceFileTypeSymlink:
		targetStateEntryFunc = func() (TargetStateEntry, error) {
			linknameBytes, err := lazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fa.Template {
				linknameBytes, err = s.ExecuteTemplateData(sourcePath, linknameBytes)
				if err != nil {
					return nil, err
				}
			}
			return &TargetStateSymlink{
				lazyLinkname: newLazyLinkname(string(bytes.TrimSpace(linknameBytes))),
			}, nil
		}
	default:
		panic(fmt.Sprintf("%s: unsupported type", string(fa.Type)))
	}

	return &SourceStateFile{
		lazyContents:         lazyContents,
		path:                 sourcePath,
		Attributes:           fa,
		targetStateEntryFunc: targetStateEntryFunc,
	}
}

// sortedTargetNames returns all of s's target names in order.
func (s *SourceState) sortedTargetNames() []string {
	targetNames := make([]string, 0, len(s.entries))
	for targetName := range s.entries {
		targetNames = append(targetNames, targetName)
	}
	sort.Strings(targetNames)
	return targetNames
}

func (s *SourceState) sourceStateEntry(system System, destPath string, info os.FileInfo, options *AddOptions) (SourceStateEntry, error) {
	destStateEntry, err := NewDestStateEntry(system, destPath)
	if err != nil {
		return nil, err
	}
	if !options.Include.IncludeDestStateEntry(destStateEntry) {
		return nil, nil
	}
	// FIXME create parents
	sourcePath := "" // FIXME
	switch destStateEntry := destStateEntry.(type) {
	case *DestStateAbsent:
		return nil, fmt.Errorf("%s: not found", destPath)
	case *DestStateDir:
		return &SourceStateDir{
			path: sourcePath,
			Attributes: DirAttributes{
				Name:    info.Name(),
				Exact:   options.Exact,
				Private: POSIXFileModes && info.Mode()&os.ModePerm&0o77 == 0,
			},
		}, nil
	case *DestStateFile:
		contents, err := destStateEntry.Contents()
		if err != nil {
			return nil, err
		}
		if options.AutoTemplate {
			contents = autoTemplate(contents, s.TemplateData())
		}
		return &SourceStateFile{
			path: sourcePath,
			Attributes: FileAttributes{
				Name:       info.Name(),
				Type:       SourceFileTypeFile,
				Empty:      options.Empty,
				Encrypted:  options.Encrypt,
				Executable: POSIXFileModes && info.Mode()&os.ModePerm&0o111 != 0,
				Private:    POSIXFileModes && info.Mode()&os.ModePerm&0o77 == 0,
				Template:   options.Template || options.AutoTemplate,
			},
			lazyContents: &lazyContents{
				contents: contents,
			},
		}, nil
	case *DestStateSymlink:
		linkname, err := destStateEntry.Linkname()
		if err != nil {
			return nil, err
		}
		contents := []byte(linkname)
		if options.AutoTemplate {
			contents = autoTemplate(contents, s.TemplateData())
		}
		return &SourceStateFile{
			path: sourcePath, // FIXME
			Attributes: FileAttributes{
				Name:     info.Name(),
				Type:     SourceFileTypeSymlink,
				Template: options.Template || options.AutoTemplate,
			},
			lazyContents: &lazyContents{
				contents: contents,
			},
		}, nil
	default:
		panic(fmt.Sprintf("%T: unsupported type", destStateEntry))
	}
}

// getTargetDirName returns the target directory name of sourceDirName.
func getTargetDirName(sourceDirName string) string {
	sourceNames := strings.Split(sourceDirName, PathSeparatorStr)
	targetNames := make([]string, 0, len(sourceNames))
	for _, sourceName := range sourceNames {
		da := parseDirAttributes(sourceName)
		targetNames = append(targetNames, da.Name)
	}
	return strings.Join(targetNames, PathSeparatorStr)
}
