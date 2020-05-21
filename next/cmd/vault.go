package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

type vaultConfig struct {
	Command string
}

var vaultCache = make(map[string]interface{})

func init() {
	config.addTemplateFunc("vault", config.vaultFunc)
}

func (c *Config) vaultFunc(key string) interface{} {
	if data, ok := vaultCache[key]; ok {
		return data
	}
	name := c.Vault.Command
	args := []string{"kv", "get", "-format=json", key}
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
	vaultCache[key] = data
	return data
}
