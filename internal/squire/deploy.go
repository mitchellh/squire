package squire

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"io/ioutil"
)

type DeployOptions struct {
	// SQL is the SQL to apply to the target. If you're resetting, you'll
	// want to clear the database and apply the Schema. If you're deploying,
	// you'll want to set this to the diff output. If this is nil, then the
	// default schema will be read.
	SQL io.Reader

	// Target is the target to apply the SQL to. For dev this could be
	// a local container, for production it could be a real remote connection.
	Target *sql.DB
}

// Deploy applies the given SQL to the target database instance.
func (s *Squire) Deploy(ctx context.Context, opts *DeployOptions) error {
	L := s.logger.Named("deploy")
	db := opts.Target

	if opts.SQL == nil {
		L.Info("schema not given, generating")
		var buf bytes.Buffer
		if err := s.Schema(&SchemaOptions{Output: &buf}); err != nil {
			L.Error("error generating schema", "err", err)
			return err
		}

		opts.SQL = &buf
	}

	// Load the SQL into memory.
	sqlbs, err := ioutil.ReadAll(opts.SQL)
	if err != nil {
		return nil
	}

	// Execute it.
	_, err = db.ExecContext(ctx, string(sqlbs))
	if err != nil {
		return err
	}

	return nil
}
