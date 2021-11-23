//go:build !windows
// +build !windows

package exec

import (
	"golang.org/x/sys/unix"
)

func realExec(argv0 string, argv []string, envv []string) error {
	return unix.Exec(argv0, argv, envv)
}
