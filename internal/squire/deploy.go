package squire

import (
	"context"
	"database/sql"
	"io"
	"io/ioutil"
)

type DeployOptions struct {
	// SQL is the SQL to apply to the target. If you're resetting, you'll
	// want to clear the database and apply the Schema. If you're deploying,
	// you'll want to set this to the diff output.
	SQL io.Reader

	// Target is the target to apply the SQL to. For dev this could be
	// a local container, for production it could be a real remote connection.
	Target *sql.DB
}

// Deploy applies the given SQL to the target database instance.
func (s *Squire) Deploy(ctx context.Context, opts *DeployOptions) error {
	db := opts.Target

	// Load the SQL into memory.
	sqlbs, err := ioutil.ReadAll(opts.SQL)
	if err != nil {
		return nil
	}

	// Execute it.
	_, err = db.ExecContext(ctx, string(sqlbs))
	if err != nil {
		return nil
	}

	return nil
}
