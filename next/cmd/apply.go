package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var applyCmd = &cobra.Command{
	Use:     "apply [targets...]",
	Short:   "Update the destination directory to match the target state",
	Long:    mustGetLongHelp("apply"),
	Example: getExample("apply"),
	RunE:    config.runApplyCmd,
	Annotations: map[string]string{
		modifiesDestinationDirectory: "true",
	},
}

type applyCmdConfig struct {
	include   *chezmoi.IncludeSet
	recursive bool
}

func init() {
	rootCmd.AddCommand(applyCmd)

	persistentFlags := applyCmd.PersistentFlags()
	persistentFlags.VarP(config.apply.include, "include", "i", "include entry types")
	persistentFlags.BoolVarP(&config.apply.recursive, "recursive", "r", config.apply.recursive, "recursive")

	markRemainingZshCompPositionalArgumentsAsFiles(applyCmd, 1)
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	return c.applyArgs(c.destSystem, c.DestDir, args, c.apply.include, c.apply.recursive)
}
