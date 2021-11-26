package squire

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/mitchellh/squire/internal/dbcontainer"
)

type DumpOptions struct {
	// TargetURI is the PostgreSQL connection address to dump.
	TargetURI string

	// Output is where the final dump is written. If this is not set, it
	// defaults to os.Stdout
	Output io.Writer
}

// Dump outputs the pg_dump for the given container.
func (s *Squire) Dump(ctx context.Context, opts *DumpOptions) error {
	L := s.logger.Named("dump")
	L.Info("starting dump")

	// Before anything else, verify we have pg_dump cause that is a dep
	pgdPath, err := exec.LookPath("pg_dump")
	if err != nil {
		return errors.WithDetail(
			errors.Newf("pg_dump could not be found: %w", err),
			strings.TrimSpace(errPGDumpNotFound),
		)
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// If the target URI is not specified, then we're using the
	// primary dev container. This must exist prior to diffing.
	if opts.TargetURI == "" {
		L.Debug("no target URI, will diff against dev container")
		ctr, err := s.Container()
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}

		st, err := ctr.Status(ctx)
		if st.State != dbcontainer.Running {
			return errors.WithDetail(
				errors.New("dev container to diff against is not running"),
				strings.TrimSpace(errDiffNotRunning),
			)
		}

		opts.TargetURI = ctr.ConnURI()
	}

	targetURI := opts.TargetURI
	L.Info("dumping", "target", targetURI)

	// Our args
	args := []string{
		"--no-comments",
		"-s", // schema only
		targetURI,
	}

	// Run
	cmd := exec.CommandContext(ctx, pgdPath, args...)
	cmd.Stdout = opts.Output
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

const (
	errPGDumpNotFound = `
The program "pg_dump" could not be found. pg_dump is required for generating
a schema dump. Schema dumps are used to verify that diffs are accurate.
Please install "pg_dump" and try again. The "pg_dump" program is usually shipped
with PostgreSQL.
`
)
