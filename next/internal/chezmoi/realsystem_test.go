package chezmoi

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"
)

var _ System = &RealSystem{}

func TestRealSystemGlob(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user": map[string]interface{}{
			"bar":            "",
			"baz":            "",
			"foo":            "",
			"dir/bar":        "",
			"dir/foo":        "",
			"dir/subdir/foo": "",
		},
	})
	require.NoError(t, err)
	defer cleanup()

	s := NewRealSystem(fs, newTestPersistentState())
	for _, tc := range []struct {
		pattern         string
		expectedMatches []string
	}{
		{
			pattern: "/home/user/foo",
			expectedMatches: []string{
				"/home/user/foo",
			},
		},
		{
			pattern: "/home/user/**/foo",
			expectedMatches: []string{
				"/home/user/dir/foo",
				"/home/user/dir/subdir/foo",
				"/home/user/foo",
			},
		},
		{
			pattern: "/home/user/**/ba*",
			expectedMatches: []string{
				"/home/user/bar",
				"/home/user/baz",
				"/home/user/dir/bar",
			},
		},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			actualMatches, err := s.Glob(tc.pattern)
			require.NoError(t, err)
			sort.Strings(actualMatches)
			assert.Equal(t, tc.expectedMatches, PathsToSlashes(actualMatches))
		})
	}
}

func newTestRealSystem(fs vfs.FS) *RealSystem {
	return NewRealSystem(fs, newTestPersistentState())
}
