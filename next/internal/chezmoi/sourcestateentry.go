package chezmoi

// A SourceStateEntry represents the state of an entry in the source state.
type SourceStateEntry interface {
	Evaluate() error
	Path() string
	TargetStateEntry() (TargetStateEntry, error)
	Write(s System) error
}

// A SourceStateDir represents the state of a directory in the source state.
type SourceStateDir struct {
	Attributes       DirAttributes
	path             string
	targetStateEntry TargetStateEntry
}

// A SourceStateFile represents the state of a file in the source state.
type SourceStateFile struct {
	*lazyContents
	Attributes           FileAttributes
	path                 string
	targetStateEntryFunc func() (TargetStateEntry, error)
	targetStateEntry     TargetStateEntry
	targetStateEntryErr  error
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateDir) Evaluate() error {
	return nil
}

// Path returns s's path.
func (s *SourceStateDir) Path() string {
	return s.path
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateDir) TargetStateEntry() (TargetStateEntry, error) {
	return s.targetStateEntry, nil
}

// Write writes s to sourceStateDir.
func (s *SourceStateDir) Write(sourceStateDir System) error {
	return sourceStateDir.Mkdir(s.path, 0o777)
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateFile) Evaluate() error {
	_, err := s.ContentsSHA256()
	return err
}

// Path returns s's path.
func (s *SourceStateFile) Path() string {
	return s.path
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateFile) TargetStateEntry() (TargetStateEntry, error) {
	if s.targetStateEntryFunc != nil {
		s.targetStateEntry, s.targetStateEntryErr = s.targetStateEntryFunc()
		s.targetStateEntryFunc = nil
	}
	return s.targetStateEntry, s.targetStateEntryErr
}

// Write writes s to sourceStateDir.
func (s *SourceStateFile) Write(sourceStateDir System) error {
	contents, err := s.Contents()
	if err != nil {
		return err
	}
	return sourceStateDir.WriteFile(s.path, contents, 0o666)
}
