package cmd

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var chattrCmd = &cobra.Command{
	Use:     "chattr attributes targets...",
	Args:    cobra.MinimumNArgs(2),
	Short:   "Change the attributes of a target in the source state",
	Long:    mustGetLongHelp("chattr"),
	Example: getExample("chattr"),
	RunE:    config.makeRunEWithSourceState(config.runChattrCmd),
	Annotations: map[string]string{
		modifiesSourceDirectory: "true",
	},
}

type boolModifier int

const (
	boolModifierSet            boolModifier = 1
	boolModifierLeaveUnchanged boolModifier = 0
	boolModifierClear          boolModifier = -1
)

type attributesModifier struct {
	empty      boolModifier
	encrypted  boolModifier
	exact      boolModifier
	executable boolModifier
	private    boolModifier
	template   boolModifier
}

func init() {
	rootCmd.AddCommand(chattrCmd)

	attributes := []string{
		"empty", "e",
		"encrypted",
		"exact",
		"executable", "x",
		"private", "p",
		"template", "t",
	}
	words := make([]string, 0, 4*len(attributes))
	for _, attribute := range attributes {
		words = append(words, attribute, "-"+attribute, "+"+attribute, "no"+attribute)
	}
	if err := chattrCmd.MarkZshCompPositionalArgumentWords(1, words...); err != nil {
		panic(err)
	}
	markRemainingZshCompPositionalArgumentsAsFiles(chattrCmd, 2)
}

func (c *Config) runChattrCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	// FIXME should the core functionality of chattr move to chezmoi.SourceState?

	am, err := parseAttributesModifier(args[0])
	if err != nil {
		return err
	}

	targetNames, err := c.getTargetNames(sourceState, args[1:], getTargetNamesOptions{
		mustBeInSourceState: true,
		recursive:           false,
	})
	if err != nil {
		return err
	}

	// Sort targets in reverse so we update children before their parent
	// directories.
	sort.Sort(sort.Reverse(sort.StringSlice(targetNames)))

	for _, targetName := range targetNames {
		sourceStateEntry := sourceState.MustEntry(targetName)
		sourceName := strings.TrimPrefix(sourceStateEntry.Path(), c.SourceDir+chezmoi.PathSeparatorStr)
		parentDirName, baseName := path.Split(sourceName)
		switch sourceStateEntry := sourceStateEntry.(type) {
		case *chezmoi.SourceStateDir:
			if newBaseName := am.modifyDirAttributes(sourceStateEntry.Attributes).BaseName(); newBaseName != baseName {
				newSourcePath := path.Join(c.SourceDir, parentDirName, newBaseName)
				if err := c.sourceSystem.Rename(sourceStateEntry.Path(), newSourcePath); err != nil {
					return err
				}
			}
		case *chezmoi.SourceStateFile:
			// FIXME encrypted attribute changes
			// FIXME when changing encrypted attribute add new file before removing old one
			if newBaseName := am.modifyFileAttributes(sourceStateEntry.Attributes).BaseName(); newBaseName != baseName {
				newSourcePath := path.Join(c.SourceDir, parentDirName, newBaseName)
				if err := c.sourceSystem.Rename(sourceStateEntry.Path(), newSourcePath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (m boolModifier) modify(b bool) bool {
	switch m {
	case boolModifierSet:
		return true
	case boolModifierLeaveUnchanged:
		return b
	case boolModifierClear:
		return false
	default:
		panic(fmt.Sprintf("%d: unknown bool modifier", m))
	}
}

func parseAttributesModifier(s string) (*attributesModifier, error) {
	am := &attributesModifier{}
	for _, modifierStr := range strings.Split(s, ",") {
		modifierStr = strings.TrimSpace(modifierStr)
		if modifierStr == "" {
			continue
		}
		var bm boolModifier
		var attribute string
		switch {
		case modifierStr[0] == '-':
			bm = boolModifierClear
			attribute = modifierStr[1:]
		case modifierStr[0] == '+':
			bm = boolModifierSet
			attribute = modifierStr[1:]
		case strings.HasPrefix(modifierStr, "no"):
			bm = boolModifierClear
			attribute = modifierStr[2:]
		default:
			bm = boolModifierSet
			attribute = modifierStr
		}
		switch attribute {
		case "empty", "e":
			am.empty = bm
		case "encrypted":
			am.encrypted = bm
		case "exact":
			am.exact = bm
		case "executable", "x":
			am.executable = bm
		case "private", "p":
			am.private = bm
		case "template", "t":
			am.template = bm
		default:
			return nil, fmt.Errorf("%s: unknown attribute", attribute)
		}
	}
	return am, nil
}

func (am *attributesModifier) modifyDirAttributes(da chezmoi.DirAttributes) chezmoi.DirAttributes {
	return chezmoi.DirAttributes{
		Name:    da.Name,
		Exact:   am.exact.modify(da.Exact),
		Private: am.private.modify(da.Private),
	}
}

func (am *attributesModifier) modifyFileAttributes(fa chezmoi.FileAttributes) chezmoi.FileAttributes {
	switch fa.Type {
	case chezmoi.SourceFileTypeFile:
		return chezmoi.FileAttributes{
			Name:       fa.Name,
			Type:       chezmoi.SourceFileTypeFile,
			Empty:      am.empty.modify(fa.Empty),
			Encrypted:  am.encrypted.modify(fa.Encrypted),
			Executable: am.executable.modify(fa.Executable),
			Private:    am.private.modify(fa.Private),
			Template:   am.template.modify(fa.Template),
		}
	case chezmoi.SourceFileTypeSymlink:
		return chezmoi.FileAttributes{
			Name:     fa.Name,
			Type:     chezmoi.SourceFileTypeSymlink,
			Template: am.template.modify(fa.Template),
		}
	default:
		panic(fmt.Sprintf("%d: unknown source file type", fa.Type))
	}
}
