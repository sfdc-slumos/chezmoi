package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var addCmd = &cobra.Command{
	Use:     "add targets...",
	Aliases: []string{"manage"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Add an existing file, directory, or symlink to the source state",
	Long:    mustGetLongHelp("add"),
	Example: getExample("add"),
	RunE:    config.makeRunEWithSourceState(config.runAddCmd),
	Annotations: map[string]string{
		modifiesSourceDirectory: "true",
	},
}

type addCmdConfig struct {
	autoTemplate bool
	empty        bool
	encrypt      bool
	exact        bool
	include      *chezmoi.IncludeSet
	recursive    bool
	template     bool
}

func init() {
	rootCmd.AddCommand(addCmd)

	persistentFlags := addCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.add.autoTemplate, "autotemplate", "a", config.add.autoTemplate, "auto generate the template when adding files as templates")
	persistentFlags.BoolVarP(&config.add.empty, "empty", "e", config.add.empty, "add empty files")
	persistentFlags.BoolVar(&config.add.encrypt, "encrypt", config.add.encrypt, "encrypt files")
	persistentFlags.BoolVarP(&config.add.exact, "exact", "x", config.add.exact, "add directories exactly")
	persistentFlags.BoolVarP(&config.add.recursive, "recursive", "r", config.add.recursive, "recursive")
	persistentFlags.BoolVarP(&config.add.template, "template", "T", config.add.template, "add files as templates")
}

func (c *Config) runAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	destPathInfos, err := c.getDestPathInfos(args, c.add.recursive)
	if err != nil {
		return err
	}

	return sourceState.Add(c.sourceSystem, c.DestDir, destPathInfos, &chezmoi.AddOptions{
		AutoTemplate: c.add.autoTemplate,
		Empty:        c.add.empty,
		Encrypt:      c.add.encrypt,
		Exact:        c.add.exact,
		Include:      c.add.include,
		Template:     c.add.template,
	})
}
