// +build !windows

package cmd

import (
	"io"
	"syscall"
)

const eolStr = "\n"

// enableVirtualTerminalProcessing does nothing.
func enableVirtualTerminalProcessing(w io.Writer) error {
	return nil
}

func getUmask() int {
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	return umask
}
