//go:build windows

package exec

import (
	"os"
	"os/exec"
)

func realExec(argv0 string, argv []string, envv []string) error {
	// Since Windows doesn't support fork/exec, we just run the process
	// and keep our parent process around. This isn't identical by any means
	// but for our usage it generally works.
	cmd := exec.Command(argv0, argv...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = envv
	return cmd.Run()
}
