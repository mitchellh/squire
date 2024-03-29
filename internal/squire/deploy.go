package squire

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
)

type DeployOptions struct {
	// SQL is the SQL to apply to the target. If you're resetting, you'll
	// want to clear the database and apply the Schema. If you're deploying,
	// you'll want to set this to the diff output. If this is nil, then the
	// default schema will be read.
	SQL io.Reader

	// Target is the target to apply the SQL to. For dev this could be
	// a local container, for production it could be a real remote connection.
	// If this is set, it takes priority over TargetURI.
	Target *sql.DB

	// TargetURI is the target to apply to the SQL to. This is used if
	// the Target is NOT set.
	TargetURI string
}

// Deploy applies the given SQL to the target database instance.
func (s *Squire) Deploy(ctx context.Context, opts *DeployOptions) error {
	L := s.logger.Named("deploy")

	if opts.SQL == nil {
		L.Info("schema not given, generating")
		var buf bytes.Buffer
		if err := s.Schema(&SchemaOptions{Output: &buf}); err != nil {
			L.Error("error generating schema", "err", err)
			return err
		}

		opts.SQL = &buf
	}

	if opts.Target == nil && opts.TargetURI != "" {
		L.Info("connecting to the DB")
		db, err := sql.Open("pgx", opts.TargetURI)
		if err != nil {
			return err
		}
		defer db.Close()

		// Wait for the connection to become ready
		err = backoff.Retry(func() error {
			return db.Ping()
		}, backoff.WithContext(
			backoff.NewConstantBackOff(15*time.Millisecond),
			ctx,
		))
		if err != nil {
			db.Close()
			return err
		}

		opts.Target = db
	}

	db := opts.Target

	// Load the SQL into memory.
	sqlbs, err := ioutil.ReadAll(opts.SQL)
	if err != nil {
		return nil
	}

	// Execute it.
	_, err = db.ExecContext(ctx, string(sqlbs))
	if err != nil {
		L.Error("error executing SQL", "err", err)

		// If this isn't a pgconn error then just return
		pgerr := &pgconn.PgError{}
		if !errors.As(err, &pgerr) {
			return err
		}

		// JSON-encode the error for now so we provide all information.
		// In the future, I want to be able to show a helpful pointer to
		// a specific context in the schema.
		human, encodeErr := json.MarshalIndent(&pgerr, "", "\t")
		if encodeErr != nil {
			// If we failed to encode, just return the original error.
			return err
		}

		// Try to find the column/line.
		// NOTE(mitchellh): In the future, we should show a clang-style
		// context so we show the SQL directly within the error message.
		line, col := positionToLineCol(sqlbs, pgerr.Position)

		// If it is a pgconn error, we want to make the output more helpful.
		return errors.Mark(errors.WithDetailf(
			errors.New(err.Error()),
			strings.TrimSpace(errDetailSqlExec),
			line, col, string(human),
		), err)
	}

	return nil
}

// positionToLineCol converts a position in character count (not byte count)
// as reported by a PostgreSQL error to a line/column for friendlier
// human output.
func positionToLineCol(src []byte, pos int32) (int, int) {
	line := 1
	col := 0

	// Go ranges over characters
	for i, c := range string(src) {
		// If we hit a newline, reset our counters
		if c == '\n' {
			line++
			col = 0
			continue
		}

		// Inc our column count
		col++

		// If we haven't reached the position yet, continue.
		if i < int(pos) {
			continue
		}

		// We found it!
		return line, col
	}

	// Never found, should not happen.
	return 0, 0
}

const (
	errDetailSqlExec = `
There was an error while executing SQL. The error happened around the
line and column shown below. Please see the schema.sql in your sql directory
to find the error.

  Line:   %[1]d
  Column: %[2]d

IMPORTANT: the line and column above does not refer to your single SQL files,
it typically refers to the compiled schema. Squire saves the full compiled
schema to your sql dir named "schema.sql". Please check that file.

For extra information, the full PostgreSQL error structure is shown below:

%[3]s
`
)
