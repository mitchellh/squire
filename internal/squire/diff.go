package squire

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"

	"github.com/mitchellh/squire/internal/dbcontainer"
	"github.com/mitchellh/squire/internal/pkg/stdcapture"
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

	// Verbose will output debug information from the diff invocation.
	Verbose bool

	// Verify verifies that the diff is complete by dumping the target
	// database, applying the diff, and then dumping againt to verify
	// it is equivalent to a reset dump. This isn't fully reliable, but
	// can be used as an additional check.
	Verify bool
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

	// We need to capture stdout/stderr because the compose API doesn't
	// allow configurable output streams.
	err = stdcapture.SuccessOnly(ioutil.Discard, ioutil.Discard, func() error {
		return source.Up(ctx)
	})
	if err != nil {
		return errors.WithDetail(
			errors.Newf("error starting source container: %w", err),
			strings.TrimSpace(errCreatingSource),
		)
	}
	defer func() {
		err = stdcapture.SuccessOnly(ioutil.Discard, ioutil.Discard, func() error {
			return source.Down(ctx)
		})
		if err != nil {
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

	// Our args
	args := []string{
		"--source-dbname", sourceURI,
		"--target-dbname", targetURI,
	}
	if opts.Verbose {
		args = append(args, "-vv")
	}

	// Write to our output, but if we're verifying we also need to store the diff.
	var diff bytes.Buffer
	output := opts.Output
	if opts.Verify {
		output = io.MultiWriter(output, &diff)
	}

	// Run pgquarrel
	cmd := exec.CommandContext(ctx, pgqPath, args...)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return err
	}

	// If we're not verifying, we're done
	if !opts.Verify {
		L.Debug("no verify, diff complete")
		return nil
	}

	// For verification, we do the following:
	// 1. Dump the target DB
	// 2. Reset using the dump in our temp container
	// 3. Apply the diff we generated
	// 4. Dump the final DB from our temp container
	// 5. Reset the temp container to our full schema
	// 6. Dump the full schema
	// 5. Compare dumps

	L.Info("verification requested, starting")
	L.Debug("dumping source database (full schema)")
	var dumpExpected bytes.Buffer
	if err := s.Dump(ctx, &DumpOptions{
		TargetURI: source.ConnURI(),
		Output:    &dumpExpected,
	}); err != nil {
		return errors.WithDetail(
			errors.Newf("error dumping source database: %w", err),
			strings.TrimSpace(errDiffDumpSource),
		)
	}

	L.Debug("dumping target database to use for reset")
	var dumpActual bytes.Buffer
	if err := s.Dump(ctx, &DumpOptions{
		TargetURI: targetURI,
		Output:    &dumpActual,
	}); err != nil {
		return errors.WithDetail(
			errors.Newf("error dumping target database: %w", err),
			strings.TrimSpace(errDiffDump),
		)
	}

	// Reset our temporary DB with the dump
	L.Debug("resetting the source container with our target dump")
	if err := s.Reset(ctx, &ResetOptions{
		Container: source,
		Schema:    bytes.NewReader(dumpActual.Bytes()),
	}); err != nil {
		return errors.WithDetail(
			errors.Newf("error applying schema to source container: %w", err),
			strings.TrimSpace(errCreatingSource),
		)
	}

	// Apply the diff
	if diff.Len() > 0 {
		L.Debug("applying the diff to the target database")
		if err := s.Deploy(ctx, &DeployOptions{
			SQL:       bytes.NewReader(diff.Bytes()),
			TargetURI: source.ConnURI(),
		}); err != nil {
			return errors.WithDetail(
				errors.Newf("error verifying diff: %w", err),
				strings.TrimSpace(errDiffVerifyApply),
			)
		}
	} else {
		L.Debug("verification with empty diff")
	}

	// Dump the temporary database again
	dumpActual.Reset()
	if err := s.Dump(ctx, &DumpOptions{
		TargetURI: source.ConnURI(),
		Output:    &dumpActual,
	}); err != nil {
		return errors.WithDetail(
			errors.Newf("error dumping verification database: %w", err),
			strings.TrimSpace(errDiffDumpSource),
		)
	}

	// Compare the diffs
	L.Debug("diffing the two dumps")
	if err := s.diffDumps(dumpExpected.Bytes(), dumpActual.Bytes()); err != nil {
		return err
	}

	L.Info("verification passed")
	return nil
}

// diffDumps diffs the two dumps. If they do not match, an error is returned
// which contains a text diff.
//
// This is NOT exact and false positives AND negatives can exist. pg_dump
// is non-deterministic and the way pgquarrel creates a diff can create differing
// dumps even if the schema is functionally equivalent. Despite this, dump
// diffing provides an extra layer of check.
func (s *Squire) diffDumps(a, b []byte) error {
	// buildLines breaks up our dump into individual lines. Empty lines
	// and comments are stripped.
	buildLines := func(v []byte) ([]string, error) {
		var result []string

		scanner := bufio.NewScanner(bytes.NewReader(v))
		for scanner.Scan() {
			txt := strings.TrimSpace(scanner.Text())

			// Ignore blank and comments
			if txt == "" || strings.HasPrefix(txt, "--") {
				continue
			}

			// Trim trailing commas so that orders in CREATE blocks don't matter
			txt = strings.TrimRight(txt, ",")

			result = append(result, txt)
		}

		return result, scanner.Err()
	}

	// Get our lines
	linesA, err := buildLines(a)
	if err != nil {
		return err
	}
	linesB, err := buildLines(b)
	if err != nil {
		return err
	}

	// Sort the lines
	sort.Strings(linesA)
	sort.Strings(linesB)

	// We just use reflect.DeepEqual since its a basic []string.
	if reflect.DeepEqual(linesA, linesB) {
		return nil
	}

	// Not equal, create a diff.
	aString := string(a)
	bString := string(b)
	edits := myers.ComputeEdits(span.URIFromPath("expected.sql"), aString, bString)
	diff := fmt.Sprint(gotextdiff.ToUnified("expected.sql", "actual.sql", aString, edits))

	return errors.WithDetailf(
		errors.New("verification failed, schema after apply does not match"),
		strings.TrimSpace(errDiffVerificationFail),
		diff,
	)
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

	errDiffDump = `
Error while dumping the target database. Squire dumps the target database
during diffing as a verification mechanism to ensure the diff is complete.
It is possible to ignore this error by disabling verification ("-verify=false"
on the "squire diff" command), but it is typically prudent to check the error
and run verification.
`

	errDiffDumpSource = `
Error while dumping the verification database. Squire dumps the verification
database to test that the generated diff would result in an equivalent schema.
It is possible to ignore this error by disabling verification ("-verify=false"
on the "squire diff" command), but it is typically prudent to check the error
and run verification.
`

	errDiffVerifyApply = `
Error while testing the generated diff on a dump of the target database.
This usually means that attempting to deploy the diff on the real target
database would fail. Inspect the error above to determine next steps.
`

	errDiffVerificationFail = `
Verification failed! During verification, Squire copies the schema of the
target database, applies the diff, and then verifies that the resulting schema
is equivalent to a full reset.

This process is NOT 100%% accurate. Both false positives and false negatives
are possible, but it helps to give an extra check during the diff. Always
scrutinize both the diff and the verification failures to ensure deploy will
do the correct thing.

The full diff of the schemas is shown below. Note that this is NOT an applyable
diff, this is a text-based schema dump diff; Squire cannot construct the
runnable SQL to reach the valid result.

The resolution to this error is usually to manually apply a small subset of
the full diff that isn't supported by Squire.

%s
`
)
