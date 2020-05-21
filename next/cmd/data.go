package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var dataCmd = &cobra.Command{
	Use:     "data",
	Args:    cobra.NoArgs,
	Short:   "Print the template data",
	Long:    mustGetLongHelp("data"),
	Example: getExample("data"),
	RunE:    config.makeRunEWithSourceState(config.runDataCmd),
}

func init() {
	rootCmd.AddCommand(dataCmd)
}

func (c *Config) runDataCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	return c.marshal(sourceState.TemplateData())
}
