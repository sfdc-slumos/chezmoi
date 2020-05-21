//go:generate go run ../internal/generate-assets -o cmd/docs.gen.go -tags=!noembeddocs -trimprefix=../ ../docs/CHANGES.md ../docs/CONTRIBUTING.md ../docs/FAQ.md ../docs/HOWTO.md ../docs/INSTALL.md ../docs/MEDIA.md ../docs/QUICKSTART.md ../docs/REFERENCE.md
//go:generate go run ../internal/generate-assets -o cmd/templates.gen.go -trimprefix=../ ../assets/templates/COMMIT_MESSAGE.tmpl
//go:generate go run ../internal/generate-helps -o cmd/helps.gen.go -i ../docs/REFERENCE.md

package main

import (
	"fmt"
	"os"

	"github.com/twpayne/chezmoi/next/cmd"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

func main() {
	if err := cmd.Execute(cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	}); err != nil {
		if s := err.Error(); s != "" {
			fmt.Fprintf(os.Stderr, "chezmoi: %s\n", s)
		}
		code := 1
		if exitCode, ok := err.(cmd.ErrExitCode); ok {
			code = int(exitCode)
		}
		os.Exit(code)
	}
}
