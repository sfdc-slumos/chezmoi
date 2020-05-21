package chezmoi

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestTargetStateEntryApplyAndEqual(t *testing.T) {
	for _, tc1 := range []struct {
		name             string
		targetStateEntry TargetStateEntry
		usesSymlinks     bool
	}{
		{
			name:             "absent",
			targetStateEntry: &TargetStateAbsent{},
		},
		{
			name: "dir",
			targetStateEntry: &TargetStateDir{
				perm: 0o755,
			},
		},
		{
			name: "file",
			targetStateEntry: &TargetStateFile{
				perm: 0o644,
				lazyContents: &lazyContents{
					contents: []byte("bar"),
				},
			},
		},
		{
			name: "file_empty",
			targetStateEntry: &TargetStateFile{
				perm: 0o644,
			},
		},
		{
			name: "file_empty_ok",
			targetStateEntry: &TargetStateFile{
				perm: 0o644,
			},
		},
		{
			name: "symlink",
			targetStateEntry: &TargetStateSymlink{
				lazyLinkname: &lazyLinkname{
					linkname: "bar",
				},
			},
			usesSymlinks: true,
		},
	} {
		t.Run(tc1.name, func(t *testing.T) {
			if tc1.usesSymlinks && !Symlinks {
				t.Skip("test uses symlinks")
			}

			for _, tc2 := range []struct {
				name         string
				root         interface{}
				usesSymlinks bool
			}{
				{
					name: "not_present",
					root: map[string]interface{}{
						"/home/user": &vfst.Dir{Perm: 0o755},
					},
				},
				{
					name: "existing_dir",
					root: map[string]interface{}{
						"/home/user/foo": &vfst.Dir{Perm: 0o755},
					},
				},
				{
					name: "existing_dir_chmod",
					root: map[string]interface{}{
						"/home/user/foo": &vfst.Dir{Perm: 0o644},
					},
				},
				{
					name: "existing_file_empty",
					root: map[string]interface{}{
						"/home/user/foo": "",
					},
				},
				{
					name: "existing_file_contents",
					root: map[string]interface{}{
						"/home/user/foo": "baz",
					},
				},
				{
					name: "existing_file_chmod",
					root: map[string]interface{}{
						"/home/user/foo": &vfst.File{
							Perm: 0o755,
						},
					},
				},
				{
					name: "existing_symlink",
					root: map[string]interface{}{
						"/home/user/bar": "",
						"/home/user/foo": &vfst.Symlink{Target: "bar"},
					},
					usesSymlinks: true,
				},
				{
					name: "existing_symlink_broken",
					root: map[string]interface{}{
						"/home/user/foo": &vfst.Symlink{Target: "bar"},
					},
					usesSymlinks: true,
				},
			} {
				t.Run(tc2.name, func(t *testing.T) {
					if tc2.usesSymlinks && !Symlinks {
						t.Skip("test uses symlinks")
					}

					testFS, cleanup, err := vfst.NewTestFS(tc2.root)
					require.NoError(t, err)
					defer cleanup()
					fs := NewRealSystem(testFS, newTestPersistentState())

					// Read the initial destination state entry from fs.
					destStateEntry, err := NewDestStateEntry(fs, "/home/user/foo")
					require.NoError(t, err)

					// Apply the target state entry.
					require.NoError(t, tc1.targetStateEntry.Apply(NewRealSystem(fs, newTestPersistentState()), destStateEntry))

					// Verify that the destination state entry matches the
					// desired state.
					vfst.RunTests(t, fs, "", vfst.TestPath("/home/user/foo", targetStateTest(t, tc1.targetStateEntry)...))

					// Read the updated destination state entry from fs and
					// verify that it is equal to the target state entry.
					newDestStateEntry, err := NewDestStateEntry(fs, "/home/user/foo")
					require.NoError(t, err)
					equal, err := tc1.targetStateEntry.Equal(newDestStateEntry)
					require.NoError(t, err)
					require.True(t, equal)
				})
			}
		})
	}
}

func targetStateTest(t *testing.T, ts TargetStateEntry) []vfst.PathTest {
	switch ts := ts.(type) {
	case *TargetStateAbsent:
		return []vfst.PathTest{
			vfst.TestDoesNotExist,
		}
	case *TargetStateDir:
		pathTests := []vfst.PathTest{
			vfst.TestIsDir,
		}
		if POSIXFileModes {
			pathTests = append(pathTests, vfst.TestModePerm(ts.perm))
		}
		return pathTests
	case *TargetStateFile:
		expectedContents, err := ts.Contents()
		require.NoError(t, err)
		pathTests := []vfst.PathTest{
			vfst.TestModeIsRegular,
			vfst.TestContents(expectedContents),
		}
		if POSIXFileModes {
			pathTests = append(pathTests, vfst.TestModePerm(ts.perm))
		}
		return pathTests
	case *TargetStateScript:
		return nil // FIXME how to verify scripts?
	case *TargetStateSymlink:
		expectedLinkname, err := ts.Linkname()
		require.NoError(t, err)
		return []vfst.PathTest{
			vfst.TestModeType(os.ModeSymlink),
			vfst.TestSymlinkTarget(expectedLinkname),
		}
	default:
		return nil
	}
}
