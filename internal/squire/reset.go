package squire

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/mitchellh/squire/internal/dbcompose"
)

// recreateDB recreates the currently selected database by issusing a
// DROP DATABASE followed by a CREATE DATABASE.
func recreateDB(ctx context.Context, cfg *dbcompose.Config) error {
	// Get our conn URL
	u, err := url.Parse(cfg.ConnURI())
	if err != nil {
		return err
	}

	// Get our prior db
	dbname := u.Path
	if dbname[0] == '/' {
		dbname = dbname[1:]
	}

	// Replace our database with "postgres" so that we can drop our database.
	u.Path = "postgres"

	// Connect,
	db, err := sql.Open("pgx", u.String())

	// Drop our database
	// NOTE: This query requires PG13+
	if _, err := db.ExecContext(ctx, `DROP DATABASE "`+dbname+`" WITH (FORCE)`); err != nil {
		return err
	}

	// Create
	if _, err := db.ExecContext(ctx, `CREATE DATABASE "`+dbname+`"`); err != nil {
		return err
	}

	return nil
}
