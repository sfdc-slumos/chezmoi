package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

type bitwardenConfig struct {
	Command string
}

var bitwardenCache = make(map[string]interface{})

func init() {
	config.addTemplateFunc("bitwarden", config.bitwardenFunc)
}

func (c *Config) bitwardenFunc(args ...string) interface{} {
	key := strings.Join(args, "\x00")
	if data, ok := bitwardenCache[key]; ok {
		return data
	}
	name := c.Bitwarden.Command
	args = append([]string{"get"}, args...)
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
	bitwardenCache[key] = data
	return data
}
