package cmd

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var executeTemplateCmd = &cobra.Command{
	Use:     "execute-template [templates...]",
	Short:   "Execute the given template(s)",
	Long:    mustGetLongHelp("execute-template"),
	Example: getExample("execute-template"),
	RunE:    config.makeRunEWithSourceState(config.runExecuteTemplateCmd),
}

type executeTemplateCmdConfig struct {
	init         bool
	promptBool   map[string]string
	promptFloat  map[string]string
	promptInt    map[string]int
	promptString map[string]string
}

func init() {
	rootCmd.AddCommand(executeTemplateCmd)

	persistentFlags := executeTemplateCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.executeTemplate.init, "init", "i", config.executeTemplate.init, "simulate chezmoi init")
	persistentFlags.StringToStringVar(&config.executeTemplate.promptBool, "promptBool", config.executeTemplate.promptBool, "simulate promptBool")
	persistentFlags.StringToStringVar(&config.executeTemplate.promptFloat, "promptFloat", config.executeTemplate.promptFloat, "simulate promptFloat")
	persistentFlags.StringToIntVar(&config.executeTemplate.promptInt, "promptInt", config.executeTemplate.promptInt, "simulate promptInt")
	persistentFlags.StringToStringVarP(&config.executeTemplate.promptString, "promptString", "p", config.executeTemplate.promptString, "simulate promptString")
}

func (c *Config) runExecuteTemplateCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	promptBool := make(map[string]bool)
	for key, valueStr := range c.executeTemplate.promptBool {
		value, err := parseBool(valueStr)
		if err != nil {
			return err
		}
		promptBool[key] = value
	}
	promptFloat := make(map[string]float64)
	for key, valueStr := range c.executeTemplate.promptFloat {
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return err
		}
		promptFloat[key] = value
	}
	if c.executeTemplate.init {
		for name, f := range map[string]interface{}{
			"promptBool": func(prompt string) bool {
				return promptBool[prompt]
			},
			"promptFloat": func(prompt string) float64 {
				return promptFloat[prompt]
			},
			"promptInt": func(prompt string) int {
				return c.executeTemplate.promptInt[prompt]
			},
			"promptString": func(prompt string) string {
				if value, ok := c.executeTemplate.promptString[prompt]; ok {
					return value
				}
				return prompt
			},
		} {
			c.templateFuncs[name] = f
		}
	}

	if len(args) == 0 {
		data, err := ioutil.ReadAll(c.stdin)
		if err != nil {
			return err
		}
		output, err := sourceState.ExecuteTemplateData("stdin", data)
		if err != nil {
			return err
		}
		return c.writeOutput(output)
	}

	output := &strings.Builder{}
	for i, arg := range args {
		result, err := sourceState.ExecuteTemplateData("arg"+strconv.Itoa(i+1), []byte(arg))
		if err != nil {
			return err
		}
		if _, err := output.Write(result); err != nil {
			return err
		}
	}
	return c.writeOutputString(output.String())
}
