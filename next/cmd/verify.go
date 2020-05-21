package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var verifyCmd = &cobra.Command{
	Use:     "verify [targets...]",
	Short:   "Exit with success if the destination state matches the target state, fail otherwise",
	Long:    mustGetLongHelp("verify"),
	Example: getExample("verify"),
	RunE:    config.runVerifyCmd,
}

type verifyCmdConfig struct {
	include   *chezmoi.IncludeSet
	recursive bool
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	persistentFlags := verifyCmd.PersistentFlags()
	persistentFlags.VarP(config.verify.include, "include", "i", "include entry types")
	persistentFlags.BoolVarP(&config.verify.recursive, "recursive", "r", config.verify.recursive, "recursive")

	markRemainingZshCompPositionalArgumentsAsFiles(verifyCmd, 1)
}

func (c *Config) runVerifyCmd(cmd *cobra.Command, args []string) error {
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	if err := c.applyArgs(dryRunSystem, c.DestDir, args, c.verify.include, c.verify.recursive); err != nil {
		return err
	}
	if dryRunSystem.Modified() {
		return ErrExitCode(1)
	}
	return nil
}
