package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/go-vfs"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var initCmd = &cobra.Command{
	Args:    cobra.MaximumNArgs(1),
	Use:     "init [repo]",
	Short:   "Setup the source directory and update the destination directory to match the target state",
	Long:    mustGetLongHelp("init"),
	Example: getExample("init"),
	RunE:    config.runInitCmd,
	Annotations: map[string]string{
		modifiesDestinationDirectory: "true", // Only if --apply. FIXME use exec instead.
		requiresSourceDirectory:      "true",
		runsCommands:                 "true",
	},
}

type initCmdConfig struct {
	apply bool
}

func init() {
	rootCmd.AddCommand(initCmd)

	persistentFlags := initCmd.PersistentFlags()
	persistentFlags.BoolVar(&config.init.apply, "apply", false, "update destination directory")
}

func (c *Config) runInitCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.run(c.SourceDir, c.Git.Command, []string{"init"})
	}

	// Clone repo into source directory.
	rawSourceDir, err := c.baseSystem.RawPath(c.SourceDir)
	if err != nil {
		return err
	}
	if err := c.run("", c.Git.Command, []string{"clone", args[0], rawSourceDir}); err != nil {
		return err
	}

	// Initialize and update submodules.
	_, err = c.baseSystem.Stat(filepath.Join(c.SourceDir, ".gitmodules"))
	switch {
	case err == nil:
		for _, args := range [][]string{
			{"submodule", "init"},
			{"submodule", "update"},
		} {
			if err := c.run(c.SourceDir, c.Git.Command, args); err != nil {
				return err
			}
		}
	case os.IsNotExist(err):
		// Do nothing.
	default:
		return err
	}

	// Find config template, execute it, and create config file.
	filename, ext, data, err := c.findConfigTemplate()
	if err != nil {
		return err
	}
	var configFileContents []byte
	if filename != "" {
		configFileContents, err = c.createConfigFile(filename, data)
		if err != nil {
			return err
		}
	}

	// If --apply is not specified, we're done.
	if !c.init.apply {
		return nil
	}

	// Reload config if it was created.
	if filename != "" {
		viper.SetConfigType(ext)
		if err := viper.ReadConfig(bytes.NewBuffer(configFileContents)); err != nil {
			return err
		}
		if err := viper.Unmarshal(c); err != nil {
			return err
		}
	}

	// Apply.
	return c.applyArgs(c.destSystem, c.DestDir, nil, chezmoi.NewIncludeSet(chezmoi.IncludeAll), true)
}

// createConfigFile creates a config file using a template and returns its
// contents.
func (c *Config) createConfigFile(filename string, data []byte) ([]byte, error) {
	funcMap := make(template.FuncMap)
	for key, value := range c.templateFuncs {
		funcMap[key] = value
	}
	for name, f := range map[string]interface{}{
		"promptBool":   c.promptBool,
		"promptFloat":  c.promptFloat,
		"promptInt":    c.promptInt,
		"promptString": c.promptString,
	} {
		funcMap[name] = f
	}

	t, err := template.New(filename).Funcs(funcMap).Parse(string(data))
	if err != nil {
		return nil, err
	}

	templateData, err := c.getDefaultTemplateData()
	if err != nil {
		return nil, err
	}

	sb := &strings.Builder{}
	if err = t.Execute(sb, map[string]interface{}{
		"chezmoi": templateData,
	}); err != nil {
		return nil, err
	}
	contents := []byte(sb.String())

	configDir := filepath.Join(c.bds.ConfigHome, "chezmoi")
	if err := vfs.MkdirAll(c.baseSystem, configDir, 0o777); err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, filename)
	if err := c.baseSystem.WriteFile(configPath, contents, 0o600); err != nil {
		return nil, err
	}

	return contents, nil
}

func (c *Config) findConfigTemplate() (string, string, []byte, error) {
	for _, ext := range viper.SupportedExts {
		filename := chezmoi.ChezmoiPrefix + "." + ext + chezmoi.TemplateSuffix
		contents, err := c.baseSystem.ReadFile(filepath.Join(c.SourceDir, filename))
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return "", "", nil, err
		}
		return "chezmoi." + ext, ext, contents, nil
	}
	return "", "", nil, nil
}

func (c *Config) promptBool(field string) bool {
	value, err := parseBool(c.promptString(field))
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptFloat(field string) float64 {
	value, err := strconv.ParseFloat(c.promptString(field), 64)
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptInt(field string) int64 {
	value, err := strconv.ParseInt(c.promptString(field), 10, 64)
	if err != nil {
		panic(err)
	}
	return value
}

func (c *Config) promptString(field string) string {
	fmt.Fprintf(c.stdout, "%s? ", field)
	value, err := bufio.NewReader(c.stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(value)
}
