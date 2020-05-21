package cmd

import (
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:     "git [args...]",
	Short:   "Run git in the source directory",
	Long:    mustGetLongHelp("git"),
	Example: getExample("git"),
	RunE:    config.runGitCmd,
	Annotations: map[string]string{
		requiresSourceDirectory: "true",
		runsCommands:            "true",
	},
}

type gitCmdConfig struct {
	Command    string
	AutoAdd    bool
	AutoCommit bool
	AutoPush   bool
}

func init() {
	rootCmd.AddCommand(gitCmd)
}

func (c *Config) runGitCmd(cmd *cobra.Command, args []string) error {
	return c.run(c.SourceDir, c.Git.Command, args)
}
