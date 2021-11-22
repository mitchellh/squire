package squire

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"io/ioutil"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/mitchellh/squire/internal/dbcontainer"
	"github.com/mitchellh/squire/internal/pkg/stdcapture"
)

//go:embed vendor/pgunit/pgunit.sql
var pgUnitSQL []byte

type TestPGUnitOptions struct {
	// Container is the primary dev container. If this is nil, the default
	// Container is used.
	Container *dbcontainer.Container

	// Callback is called with the result of calling pgUnit. This can be
	// used to inspect the results or render in any way. If this is nil,
	// the results are discarded.
	Callback func(*sql.Rows) error
}

// TestPGUnit creates a new test database with the raw schema and then runs
// the tests against it. This automatically installs pgUnit and runs all
// tests.
func (s *Squire) TestPGUnit(ctx context.Context, opts *TestPGUnitOptions) error {
	L := s.logger.Named("test")

	var err error
	if opts.Container == nil {
		opts.Container, err = s.Container()
		if err != nil {
			return err
		}
	}
	if opts.Callback == nil {
		opts.Callback = func(*sql.Rows) error { return nil }
	}

	// We need to create a temporary container to reset onto for the
	// diffing process.
	L.Debug("cloning and launching test container")
	ctr, err := opts.Container.Clone("test")
	if err != nil {
		return errors.WithDetail(
			errors.Newf("error creating test container: %w", err),
			strings.TrimSpace(errCreatingTestContainer),
		)
	}

	// We need to capture stdout/stderr because the compose API doesn't
	// allow configurable output streams.
	err = stdcapture.SuccessOnly(ioutil.Discard, ioutil.Discard, func() error {
		return ctr.Up(ctx)
	})
	if err != nil {
		return errors.WithDetail(
			errors.Newf("error starting test container: %w", err),
			strings.TrimSpace(errCreatingTestContainer),
		)
	}
	defer func() {
		err = stdcapture.SuccessOnly(ioutil.Discard, ioutil.Discard, func() error {
			return ctr.Down(ctx)
		})
		if err != nil {
			L.Error("error destroying test container, may still be dangling",
				"err", err)
		}
	}()

	// Build our full schema including tests
	var buf bytes.Buffer
	L.Debug("generating schema with tests")
	if err := s.Schema(&SchemaOptions{
		Output: &buf,
		Tests:  true,
	}); err != nil {
		L.Error("error generating schema", "err", err)
		return err
	}

	// Reset on our test container
	if err := s.Reset(ctx, &ResetOptions{
		Container: ctr,
		Schema:    &buf,
	}); err != nil {
		return errors.WithDetail(
			errors.Newf("error applying schema to source container: %w", err),
			strings.TrimSpace(errCreatingTestContainer),
		)
	}

	// Connect to our database
	L.Debug("connecting to the test database")
	db, err := ctr.Conn(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialize pgUnit
	L.Debug("deploying pgUnit")
	if err := s.Deploy(ctx, &DeployOptions{
		SQL:    bytes.NewReader(pgUnitSQL),
		Target: db,
	}); err != nil {
		return err
	}

	// Run tests
	rows, err := db.QueryContext(ctx, "select * from pgunit.test_run_all()")
	if err != nil {
		return err
	}
	defer rows.Close()

	return opts.Callback(rows)
}

const (
	errCreatingTestContainer = `
Squire creates a container clone to apply a clean version of your current
schema in order to run tests. We don't use the currently active dev container
because it might have additional data or manual changes applied.

The error above was received while attempting to start this test container.
Please resolve the error and try again.
`
)
