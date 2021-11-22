package squire

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/mitchellh/squire/internal/dbcontainer"
)

type DiffOptions struct {
	// Container is the primary dev container. If no target URI is specified,
	// the primary dev container is used as the target. If a target URI is
	// specified, then this is only used to create a temporary instance to
	// compare against a clean schema. If this is nil, the default Container
	// is used.
	Container *dbcontainer.Container

	// TargetURI is the PostgreSQL connection address with the current
	// "live" schema. If this is empty, then the target will be the primary
	// development container.
	TargetURI string

	// Output is where the final diff is written. If this is not set, it
	// defaults to os.Stdout
	Output io.Writer
}

// Diff creates a diff between two database instances.
func (s *Squire) Diff(ctx context.Context, opts *DiffOptions) error {
	L := s.logger.Named("diff")
	L.Info("starting diff")

	// Before anything else, verify we have pgquarrel cause that is a dep
	pgqPath, err := exec.LookPath("pgquarrel")
	if err != nil {
		return errors.WithDetail(
			errors.Newf("pgquarrel could not be found: %w", err),
			strings.TrimSpace(errPGQuarrelNotFound),
		)
	}

	if opts.Container == nil {
		opts.Container, err = s.Container()
		if err != nil {
			return err
		}
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// If the target URI is not specified, then we're using the
	// primary dev container. This must exist prior to diffing.
	if opts.TargetURI == "" {
		L.Debug("no target URI, will diff against dev container")
		st, err := opts.Container.Status(ctx)
		if err != nil {
			return err
		}

		if st.State != dbcontainer.Running {
			return errors.WithDetail(
				errors.New("dev container to diff against is not running"),
				strings.TrimSpace(errDiffNotRunning),
			)
		}

		opts.TargetURI = opts.Container.ConnURI()
	}

	// We need to create a temporary container to reset onto for the
	// diffing process.
	L.Debug("cloning and launching source container")
	source, err := opts.Container.Clone(fmt.Sprintf("diff-%d", time.Now().Unix()))
	if err != nil {
		return errors.WithDetail(
			errors.Newf("error creating source container: %w", err),
			strings.TrimSpace(errCreatingSource),
		)
	}
	if err := source.Up(ctx); err != nil {
		return errors.WithDetail(
			errors.Newf("error starting source container: %w", err),
			strings.TrimSpace(errCreatingSource),
		)
	}
	defer func() {
		if err := source.Down(ctx); err != nil {
			L.Error("error destroying source container, may still be dangling",
				"err", err)
		}
	}()

	// Reset on our source
	if err := s.Reset(ctx, &ResetOptions{
		Container: source,
	}); err != nil {
		return errors.WithDetail(
			errors.Newf("error applying schema to source container: %w", err),
			strings.TrimSpace(errCreatingSource),
		)
	}

	sourceURI := source.ConnURI()
	targetURI := opts.TargetURI
	L.Info("diffing", "source", sourceURI, "target", targetURI)

	// Run pgquarrel
	cmd := exec.CommandContext(ctx, pgqPath,
		"--source-dbname", sourceURI,
		"--target-dbname", targetURI,
	)
	cmd.Stdout = opts.Output
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

const (
	errDiffNotRunning = `
A diff was requested against the currently deployed dev container, but
the dev container is not currently running. Please start the dev container
with "squire up".

If instead you meant to diff against another database, please specify
the proper flags or configuration to the diff command. See "squire diff -h"
for more help.
`

	errCreatingSource = `
Squire creates a container clone to apply a clean version of your current
schema in order to create the diff. We don't use the currently active
dev container because it might have additional data or manual changes applied
or it might be the target database for the diff.

The error above was received while attempting to start this source container
for diffing. Please resolve the error and try again.
`

	errPGQuarrelNotFound = `
The program "pgquarrel" could not be found. pgquarrel is required for diffing
schemas. Please install pgquarrel prior to continuing.

https://github.com/eulerto/pgquarrel
`
)
