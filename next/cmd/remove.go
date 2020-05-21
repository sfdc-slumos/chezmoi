package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var removeCmd = &cobra.Command{
	Use:     "remove targets...",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a target from the source state and the destination directory",
	Long:    mustGetLongHelp("remove"),
	Example: getExample("remove"),
	RunE:    config.makeRunEWithSourceState(config.runRemoveCmd),
	Annotations: map[string]string{
		modifiesDestinationDirectory: "true",
		modifiesSourceDirectory:      "true",
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	markRemainingZshCompPositionalArgumentsAsFiles(removeCmd, 1)
}

func (c *Config) runRemoveCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetNames, err := c.getTargetNames(sourceState, args, getTargetNamesOptions{
		recursive:           false,
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	for _, targetName := range targetNames {
		destPath := path.Join(c.DestDir, targetName)
		sourcePath := sourceState.MustEntry(targetName).Path()
		if !c.force {
			choice, err := c.prompt(fmt.Sprintf("Remove %s and %s", destPath, sourcePath), "ynqa")
			if err != nil {
				return err
			}
			switch choice {
			case 'y':
			case 'n':
				continue
			case 'q':
				return nil
			case 'a':
				c.force = true
			}
		}
		if err := c.destSystem.RemoveAll(destPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := c.sourceSystem.RemoveAll(sourcePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
