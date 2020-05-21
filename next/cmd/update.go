package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Args:    cobra.NoArgs,
	Short:   "Pull and apply any changes",
	Long:    mustGetLongHelp("update"),
	Example: getExample("update"),
	RunE:    config.runUpdateCmd,
	Annotations: map[string]string{
		modifiesDestinationDirectory: "true",
		requiresSourceDirectory:      "true",
		runsCommands:                 "true",
	},
}

type updateCmdConfig struct {
	apply     bool
	include   *chezmoi.IncludeSet
	recursive bool
}

func init() {
	rootCmd.AddCommand(updateCmd)

	persistentFlags := updateCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.update.apply, "apply", "a", config.update.apply, "apply after pulling")
	persistentFlags.VarP(config.update.include, "include", "i", "include entry types")
	persistentFlags.BoolVarP(&config.update.recursive, "recursive", "r", config.update.recursive, "recursive")

	markRemainingZshCompPositionalArgumentsAsFiles(updateCmd, 1)
}

func (c *Config) runUpdateCmd(cmd *cobra.Command, args []string) error {
	if err := c.run(c.SourceDir, c.Git.Command, []string{"pull", "--rebase"}); err != nil {
		return err
	}

	if !c.update.apply {
		return nil
	}

	return c.applyArgs(c.destSystem, c.DestDir, args, c.update.include, c.update.recursive)
}
