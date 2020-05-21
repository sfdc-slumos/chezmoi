package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAttributesModifier(t *testing.T) {
	for _, tc := range []struct {
		s           string
		expected    *attributesModifier
		expectedErr bool
	}{
		{
			s: "empty",
			expected: &attributesModifier{
				empty: boolModifierSet,
			},
		},
		{
			s: "+empty",
			expected: &attributesModifier{
				empty: boolModifierSet,
			},
		},
		{
			s: "-empty",
			expected: &attributesModifier{
				empty: boolModifierClear,
			},
		},
		{
			s: "noempty",
			expected: &attributesModifier{
				empty: boolModifierClear,
			},
		},
		{
			s: "e",
			expected: &attributesModifier{
				empty: boolModifierSet,
			},
		},
		{
			s: "encrypted",
			expected: &attributesModifier{
				encrypted: boolModifierSet,
			},
		},
		{
			s: "executable",
			expected: &attributesModifier{
				executable: boolModifierSet,
			},
		},
		{
			s: "x",
			expected: &attributesModifier{
				executable: boolModifierSet,
			},
		},
		{
			s: "private",
			expected: &attributesModifier{
				private: boolModifierSet,
			},
		},
		{
			s: "p",
			expected: &attributesModifier{
				private: boolModifierSet,
			},
		},
		{
			s: "template",
			expected: &attributesModifier{
				template: boolModifierSet,
			},
		},
		{
			s: "t",
			expected: &attributesModifier{
				template: boolModifierSet,
			},
		},
		{
			s: "empty,+executable,noprivate,-t",
			expected: &attributesModifier{
				empty:      boolModifierSet,
				executable: boolModifierSet,
				private:    boolModifierClear,
				template:   boolModifierClear,
			},
		},
		{
			s:           "once",
			expectedErr: true,
		},
		{
			s: " empty , -private, notemplate ",
			expected: &attributesModifier{
				empty:    boolModifierSet,
				private:  boolModifierClear,
				template: boolModifierClear,
			},
		},
		{
			s: "p,,-t",
			expected: &attributesModifier{
				private:  boolModifierSet,
				template: boolModifierClear,
			},
		},
	} {
		t.Run(tc.s, func(t *testing.T) {
			actual, err := parseAttributesModifier(tc.s)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
