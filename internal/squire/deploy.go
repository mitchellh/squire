package squire

import (
	"database/sql"
	"io"
)

type DeployOptions struct {
	// SQL is the SQL to apply to the target. If you're resetting, you'll
	// want to clear the database and apply the Schema. If you're deploying,
	// you'll want to set this to the diff output.
	SQL io.Reader

	// Target is the target to apply the SQL to. For dev this could be
	// a local container, for production it could be a real remote connection.
	Target sql.Conn
}

// Deploy applies the given SQL to the target database instance.
func (s *Squire) Deploy(opts *DeployOptions) error {
	return nil
}
