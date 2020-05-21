package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var dumpCmd = &cobra.Command{
	Use:     "dump [targets...]",
	Short:   "Generate a dump of the target state",
	Long:    mustGetLongHelp("dump"),
	Example: getExample("dump"),
	RunE:    config.runDumpCmd,
}

type dumpCmdConfig struct {
	include   *chezmoi.IncludeSet
	recursive bool
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	persistentFlags := dumpCmd.PersistentFlags()
	persistentFlags.VarP(config.dump.include, "include", "i", "include entry types")
	persistentFlags.BoolVarP(&config.dump.recursive, "recursive", "r", config.dump.recursive, "recursive")

	markRemainingZshCompPositionalArgumentsAsFiles(dumpCmd, 1)
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dumpSystem := chezmoi.NewDumpSystem()
	if err := c.applyArgs(dumpSystem, "", args, c.dump.include, c.dump.recursive); err != nil {
		return err
	}
	return c.marshal(dumpSystem.Data())
}
