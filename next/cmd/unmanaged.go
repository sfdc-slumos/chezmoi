package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var unmanagedCmd = &cobra.Command{
	Use:     "unmanaged",
	Args:    cobra.NoArgs,
	Short:   "List the unmanaged files in the destination directory",
	Long:    mustGetLongHelp("unmanaged"),
	Example: getExample("unmanaged"),
	RunE:    config.makeRunEWithSourceState(config.runUnmanagedCmd),
}

func init() {
	rootCmd.AddCommand(unmanagedCmd)
}

func (c *Config) runUnmanagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	sb := &strings.Builder{}
	if err := vfs.WalkSlash(c.destSystem, c.DestDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == c.DestDir {
			return nil
		}
		targetName := strings.TrimPrefix(path, c.DestDir+chezmoi.PathSeparatorStr)
		_, managed := sourceState.Entry(targetName)
		ignored := sourceState.Ignored(targetName)
		if !managed && !ignored {
			sb.WriteString(targetName + eolStr)
		}
		if info.IsDir() && (!managed || ignored) {
			return vfs.SkipDir
		}
		return nil
	}); err != nil {
		return err
	}
	return c.writeOutputString(sb.String())
}
