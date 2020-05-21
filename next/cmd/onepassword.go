package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

type onepasswordConfig struct {
	Command string
}

var (
	onepasswordCache         = make(map[string]interface{})
	onepasswordDocumentCache = make(map[string]string)
)

func init() {
	config.addTemplateFunc("onepassword", config.onepasswordFunc)
	config.addTemplateFunc("onepasswordDocument", config.onepasswordDocumentFunc)
}

func (c *Config) onepasswordFunc(item string) interface{} {
	if data, ok := onepasswordCache[item]; ok {
		return data
	}
	name := c.Onepassword.Command
	args := []string{"get", "item", item}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
	}
	onepasswordCache[item] = data
	return data
}

func (c *Config) onepasswordDocumentFunc(item string) interface{} {
	if output, ok := onepasswordDocumentCache[item]; ok {
		return output
	}
	name := c.Onepassword.Command
	args := []string{"get", "document", item}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
	}
	onepasswordDocumentCache[item] = string(output)
	return string(output)
}
