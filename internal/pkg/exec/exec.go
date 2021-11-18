// Package exec provides a Unix-style "exec" for all operating systems.
// For operating systems that don't support exec (i.e. Windows), we
// just use `os/exec` for the duration of the child process. Kind of janky.
package exec

// Exec is like Unix syscall.Exec
func Exec(argv0 string, argv []string, envv []string) error {
	return realExec(argv0, argv, envv)
}
