package cmd

import (
	"archive/tar"
	"compress/gzip"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
)

var archiveCmd = &cobra.Command{
	Use:     "archive [targets...]",
	Short:   "Generate a tar archive of the target state",
	Long:    mustGetLongHelp("archive"),
	Example: getExample("archive"),
	RunE:    config.runArchiveCmd,
}

type archiveCmdConfig struct {
	gzip      bool
	include   *chezmoi.IncludeSet
	recursive bool
}

func init() {
	rootCmd.AddCommand(archiveCmd)

	persistentFlags := archiveCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.archive.gzip, "gzip", "z", config.archive.gzip, "compress the output with gzip")
	persistentFlags.VarP(config.archive.include, "include", "i", "include entry types")
	persistentFlags.BoolVarP(&config.archive.recursive, "recursive", "r", config.archive.recursive, "recursive")
}

func (c *Config) runArchiveCmd(cmd *cobra.Command, args []string) error {
	archive := &strings.Builder{}
	tarSystem := chezmoi.NewTARSystem(archive, tarHeaderTemplate())
	if err := c.applyArgs(tarSystem, "", args, c.archive.include, c.archive.recursive); err != nil {
		return err
	}
	if err := tarSystem.Close(); err != nil {
		return err
	}

	if !c.archive.gzip {
		return c.writeOutputString(archive.String())
	}

	gzippedArchive := &strings.Builder{}
	w := gzip.NewWriter(gzippedArchive)
	if _, err := w.Write([]byte(archive.String())); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.writeOutputString(gzippedArchive.String())
}

// tarHeaderTemplate returns a tar.Header template populated with the current
// user and time.
func tarHeaderTemplate() tar.Header {
	// Attempt to lookup the current user. Ignore errors because the default
	// zero values are reasonable.
	var (
		uid   int
		gid   int
		uname string
		gname string
	)
	if currentUser, err := user.Current(); err == nil {
		uid, _ = strconv.Atoi(currentUser.Uid)
		gid, _ = strconv.Atoi(currentUser.Gid)
		uname = currentUser.Username
		if group, err := user.LookupGroupId(currentUser.Gid); err == nil {
			gname = group.Name
		}
	}

	now := time.Now()
	return tar.Header{
		Uid:        uid,
		Gid:        gid,
		Uname:      uname,
		Gname:      gname,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
}
