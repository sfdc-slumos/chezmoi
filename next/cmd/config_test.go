package cmd

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

func TestAddTemplateFuncPanic(t *testing.T) {
	c := newTestConfig(t, nil)
	assert.NotPanics(t, func() {
		c.addTemplateFunc("func", nil)
	})
	assert.Panics(t, func() {
		c.addTemplateFunc("func", nil)
	})
}

func TestUpperSnakeCaseToCamelCase(t *testing.T) {
	for s, want := range map[string]string{
		"BUG_REPORT_URL":   "bugReportURL",
		"ID":               "id",
		"ID_LIKE":          "idLike",
		"NAME":             "name",
		"VERSION_CODENAME": "versionCodename",
		"VERSION_ID":       "versionID",
	} {
		assert.Equal(t, want, upperSnakeCaseToCamelCase(s))
	}
}

func TestValidateKeys(t *testing.T) {
	for _, tc := range []struct {
		data    interface{}
		wantErr bool
	}{
		{
			data:    nil,
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"foo":                    "bar",
				"a":                      0,
				"_x9":                    false,
				"ThisVariableIsExported": nil,
				"αβ":                     "",
			},
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"foo-foo": "bar",
			},
			wantErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar-bar": "baz",
				},
			},
			wantErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar-bar": "baz",
					},
				},
			},
			wantErr: true,
		},
	} {
		if tc.wantErr {
			assert.Error(t, validateKeys(tc.data, identifierRegexp))
		} else {
			assert.NoError(t, validateKeys(tc.data, identifierRegexp))
		}
	}
}

func newTestConfig(t *testing.T, fs vfs.FS, options ...configOption) *Config {
	c, err := newConfig(append(
		[]configOption{
			withTestFS(fs),
			withTestUser("user"),
		},
		options...,
	)...)
	require.NoError(t, err)
	return c
}

func withAddCmdConfig(add addCmdConfig) configOption {
	return func(c *Config) {
		c.add = add
	}
}

func withBaseSystem(baseSystem chezmoi.System) configOption {
	return func(c *Config) {
		c.baseSystem = baseSystem
	}
}

func withData(data map[string]interface{}) configOption {
	return func(c *Config) {
		c.Data = data
	}
}

func withDestDir(destDir string) configOption {
	return func(c *Config) {
		c.DestDir = destDir
	}
}

func withDestSystem(destSystem chezmoi.System) configOption {
	return func(c *Config) {
		c.destSystem = destSystem
	}
}

func withGenericSecretCmdConfig(genericSecretCmdConfig genericSecretConfig) configOption {
	return func(c *Config) {
		c.GenericSecret = genericSecretCmdConfig
	}
}

func withRemove(remove bool) configOption {
	return func(c *Config) {
		c.Remove = remove
	}
}

func withSourceSystem(sourceSystem chezmoi.System) configOption {
	return func(c *Config) {
		c.sourceSystem = sourceSystem
	}
}

func withStdin(stdin io.Reader) configOption {
	return func(c *Config) {
		c.stdin = stdin
	}
}

func withStdout(stdout io.WriteCloser) configOption {
	return func(c *Config) {
		c.stdout = stdout
	}
}

func withTestFS(fs vfs.FS) configOption {
	return func(c *Config) {
		c.fs = fs
	}
}

func withTestUser(username string) configOption {
	return func(c *Config) {
		homeDir := filepath.Join("/", "home", username)
		c.SourceDir = filepath.Join(homeDir, ".local", "share", "chezmoi")
		c.DestDir = homeDir
		c.Umask = 0o22
		c.bds = &xdg.BaseDirectorySpecification{
			ConfigHome: filepath.Join(homeDir, ".config"),
			DataHome:   filepath.Join(homeDir, ".local"),
			CacheHome:  filepath.Join(homeDir, ".cache"),
			RuntimeDir: filepath.Join(homeDir, ".run"),
		}
	}
}
