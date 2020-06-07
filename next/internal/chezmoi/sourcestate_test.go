package chezmoi

import (
	"os"
	"testing"
	"text/template"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestSourceStateApplyAll(t *testing.T) {
	// FIXME script tests
	// FIXME script template tests
	for _, tc := range []struct {
		name               string
		root               interface{}
		sourceStateOptions []SourceStateOption
		tests              []interface{}
	}{
		{
			name: "empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": &vfst.Dir{Perm: 0o755},
				},
			},
		},
		{
			name: "dir",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"foo": &vfst.Dir{Perm: 0o755},
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestIsDir,
					vfst.TestModePerm(0o755),
				),
			},
		},
		{
			name: "dir_exact",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "",
					},
					".local/share/chezmoi": map[string]interface{}{
						"exact_foo": &vfst.Dir{Perm: 0o755},
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestIsDir,
					vfst.TestModePerm(0o755),
				),
				vfst.TestPath("/home/user/foo/bar",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "file",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o644),
					vfst.TestContentsString("bar"),
				),
			},
		},
		{
			name: "file_remove_empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					"foo": "",
					".local/share/chezmoi": map[string]interface{}{
						"foo": "",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "file_create_empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"empty_foo": "",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o644),
					vfst.TestContentsString(""),
				),
			},
		},
		{
			name: "file_template",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"foo.tmpl": "email = {{ .email }}",
					},
				},
			},
			sourceStateOptions: []SourceStateOption{
				WithTemplateData(map[string]interface{}{
					"email": "you@example.com",
				}),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o644),
					vfst.TestContentsString("email = you@example.com"),
				),
			},
		},
		{
			name: "symlink",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"symlink_foo": "bar",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("bar"),
				),
			},
		},
		{
			name: "symlink_template",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"symlink_foo.tmpl": "bar_{{ .os }}",
					},
				},
			},
			sourceStateOptions: []SourceStateOption{
				WithTemplateData(map[string]interface{}{
					"os": "linux",
				}),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("bar_linux"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()

			system := newTestRealSystem(fs)
			sourceStateOptions := []SourceStateOption{
				WithSystem(system),
				WithSourcePath("/home/user/.local/share/chezmoi"),
			}
			sourceStateOptions = append(sourceStateOptions, tc.sourceStateOptions...)
			s := NewSourceState(sourceStateOptions...)
			require.NoError(t, s.Read())
			require.NoError(t, s.Evaluate())
			require.NoError(t, s.ApplyAll(system, "/home/user", NewIncludeSet(IncludeAll)))

			vfst.RunTests(t, fs, "", tc.tests...)
		})
	}
}

func TestSourceStateRead(t *testing.T) {
	for _, tc := range []struct {
		name                string
		root                interface{}
		sourceStateOptions  []SourceStateOption
		expectedError       string
		expectedSourceState *SourceState
	}{
		{
			name: "empty",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": &vfst.Dir{Perm: 0o755},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
			),
		},
		{
			name: "dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo": &vfst.Dir{Perm: 0o755},
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withEntries(map[string]SourceStateEntry{
					"foo": &SourceStateDir{
						path: "/home/user/.local/share/chezmoi/foo",
						Attributes: DirAttributes{
							Name: "foo",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o755,
						},
					},
				}),
			),
		},
		{
			name: "file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo": "bar",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withEntries(map[string]SourceStateEntry{
					"foo": &SourceStateFile{
						path: "/home/user/.local/share/chezmoi/foo",
						Attributes: FileAttributes{
							Name: "foo",
							Type: SourceFileTypeFile,
						},
						lazyContents: newLazyContents([]byte("bar")),
						targetStateEntry: &TargetStateFile{
							perm:         0o644,
							lazyContents: newLazyContents([]byte("bar")),
						},
					},
				}),
			),
		},
		{
			name: "duplicate_target",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo":      "bar",
					"foo.tmpl": "bar",
				},
			},
			expectedError: "foo: duplicate target (/home/user/.local/share/chezmoi/foo, /home/user/.local/share/chezmoi/foo.tmpl)",
		},
		{
			name: "duplicate_target",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo":       "bar",
					"exact_foo": &vfst.Dir{Perm: 0o755},
				},
			},
			expectedError: "foo: duplicate target (/home/user/.local/share/chezmoi/exact_foo, /home/user/.local/share/chezmoi/foo)",
		},
		{
			name: "symlink",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo": &vfst.Symlink{Target: "bar"},
				},
			},
			expectedError: "/home/user/.local/share/chezmoi/foo: unsupported file type symlink",
		},
		{
			name: "script",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"run_foo": "bar",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withEntries(map[string]SourceStateEntry{
					"foo": &SourceStateFile{
						path: "/home/user/.local/share/chezmoi/run_foo",
						Attributes: FileAttributes{
							Name: "foo",
							Type: SourceFileTypeScript,
						},
						lazyContents: newLazyContents([]byte("bar")),
						targetStateEntry: &TargetStateScript{
							name:         "foo",
							lazyContents: newLazyContents([]byte("bar")),
						},
					},
				}),
			),
		},
		{
			name: "symlink",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"symlink_foo": "bar",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withEntries(map[string]SourceStateEntry{
					"foo": &SourceStateFile{
						path: "/home/user/.local/share/chezmoi/symlink_foo",
						Attributes: FileAttributes{
							Name: "foo",
							Type: SourceFileTypeSymlink,
						},
						lazyContents: newLazyContents([]byte("bar")),
						targetStateEntry: &TargetStateSymlink{
							lazyLinkname: newLazyLinkname("bar"),
						},
					},
				}),
			),
		},
		{
			name: "file_in_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "baz",
					},
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withEntries(map[string]SourceStateEntry{
					"foo": &SourceStateDir{
						path: "/home/user/.local/share/chezmoi/foo",
						Attributes: DirAttributes{
							Name: "foo",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o755,
						},
					},
					"foo/bar": &SourceStateFile{
						path: "/home/user/.local/share/chezmoi/foo/bar",
						Attributes: FileAttributes{
							Name: "bar",
							Type: SourceFileTypeFile,
						},
						lazyContents: &lazyContents{
							contents: []byte("baz"),
						},
						targetStateEntry: &TargetStateFile{
							perm: 0o644,
							lazyContents: &lazyContents{
								contents: []byte("baz"),
							},
						},
					},
				}),
			),
		},
		{
			name: "chezmoiignore",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiignore": "README.md\n",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withIgnore(
					NewPatternSet(
						withAdd(t, "README.md", true),
					),
				),
			),
		},
		{
			name: "chezmoiignore_ignore_file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiignore": "README.md\n",
					"README.md":      "",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withIgnore(
					NewPatternSet(
						withAdd(t, "README.md", true),
					),
				),
			),
		},
		{
			name: "chezmoiremove",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiremove": "!*.txt\n",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withRemove(
					NewPatternSet(
						withAdd(t, "*.txt", false),
					),
				),
			),
		},
		{
			name: "chezmoitemplates",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoitemplates": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withTemplates(
					map[string]*template.Template{
						"foo": template.Must(template.New("foo").Parse("bar")),
					},
				),
			),
		},
		{
			name: "chezmoiversion",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiversion": "1.2.3\n",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withMinVersion(
					semver.Version{
						Major: 1,
						Minor: 2,
						Patch: 3,
					},
				),
			),
		},
		{
			name: "chezmoiversion_multiple",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiversion": "1.2.3\n",
					"foo": map[string]interface{}{
						".chezmoiversion": "2.3.4\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
				withEntries(map[string]SourceStateEntry{
					"foo": &SourceStateDir{
						path: "/home/user/.local/share/chezmoi/foo",
						Attributes: DirAttributes{
							Name: "foo",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o755,
						},
					},
				}),
				withMinVersion(
					semver.Version{
						Major: 2,
						Minor: 3,
						Patch: 4,
					},
				),
			),
		},
		{
			name: "ignore_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".ignore": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
			),
		},
		{
			name: "ignore_file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".ignore": "",
				},
			},
			expectedSourceState: NewSourceState(
				WithSourcePath("/home/user/.local/share/chezmoi"),
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()

			sourceStateOptions := []SourceStateOption{
				WithSystem(newTestRealSystem(fs)),
				WithSourcePath("/home/user/.local/share/chezmoi"),
			}
			sourceStateOptions = append(sourceStateOptions, tc.sourceStateOptions...)
			s := NewSourceState(sourceStateOptions...)
			err = s.Read()
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			require.NoError(t, s.Evaluate())
			require.NoError(t, tc.expectedSourceState.Evaluate())
			s.system = nil
			assert.Equal(t, tc.expectedSourceState, s)
		})
	}
}

func withEntries(sourceEntries map[string]SourceStateEntry) SourceStateOption {
	return func(s *SourceState) {
		s.entries = sourceEntries
	}
}

func withIgnore(ignore *PatternSet) SourceStateOption {
	return func(s *SourceState) {
		s.ignore = ignore
	}
}

func withMinVersion(minVersion semver.Version) SourceStateOption {
	return func(s *SourceState) {
		s.minVersion = minVersion
	}
}

func withRemove(remove *PatternSet) SourceStateOption {
	return func(s *SourceState) {
		s.remove = remove
	}
}

func withTemplates(templates map[string]*template.Template) SourceStateOption {
	return func(s *SourceState) {
		s.templates = templates
	}
}
