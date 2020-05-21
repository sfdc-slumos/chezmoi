package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                "chezmoi",
	Short:              "Manage your dotfiles across multiple machines, securely",
	SilenceErrors:      true,
	SilenceUsage:       true,
	PersistentPreRunE:  config.persistentPreRunRootE,
	PersistentPostRunE: config.persistentPostRunRootE,
}

const (
	doesNotRequireValidConfig    = "chezmoi_annotation_does_not_require_valid_config"
	modifiesConfigFile           = "chezmoi_annotation_modifies_config_file"
	modifiesDestinationDirectory = "chezmoi_annotation_modifies_destination_directory"
	modifiesSourceDirectory      = "chezmoi_annotation_modifies_source_directory"
	requiresConfigDirectory      = "chezmoi_annotation_requires_config_directory"
	requiresSourceDirectory      = "chezmoi_annotation_requires_source_directory"
	runsCommands                 = "chezmoi_annotation_runs_commands"
)

var (
	config  = mustNewConfig()
	initErr error
)

// An ErrExitCode indicates the the main program should exit with the given
// code.
type ErrExitCode int

func (e ErrExitCode) Error() string { return "" }

// A VersionInfo contains a version.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

func init() {
	if err := config.register(rootCmd); err != nil {
		initErr = err
		return
	}
}

// Execute executes the root command.
func Execute(v VersionInfo) error {
	if initErr != nil {
		return initErr
	}

	var versionComponents []string
	if v.Version != "" {
		var err error
		config.version, err = semver.NewVersion(strings.TrimPrefix(v.Version, "v"))
		if err != nil {
			return err
		}
		versionComponents = append(versionComponents, config.version.String())
	} else {
		versionComponents = append(versionComponents, "dev")
	}
	if v.Commit != "" {
		versionComponents = append(versionComponents, "commit "+v.Commit)
	}
	if v.Date != "" {
		versionComponents = append(versionComponents, "built at "+v.Date)
	}
	if v.BuiltBy != "" {
		versionComponents = append(versionComponents, "built by "+v.BuiltBy)
	}
	rootCmd.Version = strings.Join(versionComponents, ", ")

	return rootCmd.Execute()
}

func getAsset(name string) ([]byte, error) {
	asset, ok := assets[name]
	if !ok {
		return nil, fmt.Errorf("%s: not found", name)
	}
	return asset, nil
}

func getBoolAnnotation(cmd *cobra.Command, key string) bool {
	value, ok := cmd.Annotations[key]
	if !ok {
		return false
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}
	return boolValue
}

func getExample(command string) string {
	return helps[command].example
}

func markRemainingZshCompPositionalArgumentsAsFiles(cmd *cobra.Command, from int) {
	// As far as I can tell, there is no way to mark all remaining positional
	// arguments as files. Marking the first eight positional arguments as files
	// should be enough for everybody.
	// FIXME mark all remaining positional arguments as files
	for i := 0; i < 8; i++ {
		if err := cmd.MarkZshCompPositionalArgumentFile(from + i); err != nil {
			panic(err)
		}
	}
}

func mustGetLongHelp(command string) string {
	help, ok := helps[command]
	if !ok {
		panic(fmt.Sprintf("%s: no long help", command))
	}
	return help.long
}
