package cmd

import (
	"github.com/spf13/cobra"
)

var editConfigCmd = &cobra.Command{
	Use:     "edit-config",
	Args:    cobra.NoArgs,
	Short:   "Edit the configuration file",
	Long:    mustGetLongHelp("edit-config"),
	Example: getExample("edit-config"),
	RunE:    config.runEditConfigCmd,
	Annotations: map[string]string{
		modifiesConfigFile:      "true",
		requiresConfigDirectory: "true",
		runsCommands:            "true",
	},
}

func init() {
	rootCmd.AddCommand(editConfigCmd)
}

func (c *Config) runEditConfigCmd(cmd *cobra.Command, args []string) error {
	return c.runEditor([]string{c.configFile})
}
