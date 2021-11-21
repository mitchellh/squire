package squire

import (
	"context"
	"database/sql"
	"io"
	"net/url"

	"github.com/mitchellh/squire/internal/dbcompose"
	"github.com/mitchellh/squire/internal/dbcontainer"
)

type ResetOptions struct {
	// Container is the container to reset. Reset isn't supported in
	// non-dev mode (at the moment) because we need to recreate the whole
	// database which requires superuser. If this isn't set, the default
	// Container will be used.
	Container *dbcontainer.Container

	// Schema to apply upon reset. If this isn't set, a default schema
	// will be loaded by calling Schema.
	Schema io.Reader
}

// Reset recreates the entire database quickly by dropping the
// database, recreating, and reapplying the SQL. It doesn't recreate
// the container by default.
func (s *Squire) Reset(ctx context.Context, opts *ResetOptions) error {
	L := s.logger.Named("reset")
	L.Info("starting reset")

	var err error
	if opts.Container == nil {
		opts.Container, err = s.Container()
		if err != nil {
			return err
		}
	}

	// Recreate the database first
	L.Debug("recreating the logical database")
	if err := recreateDB(ctx, opts.Container.Config()); err != nil {
		L.Error("error recreating the db", "err", err)
		return err
	}

	// Connect
	L.Debug("connecting to database")
	db, err := opts.Container.Conn(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	// Apply the schema
	L.Debug("deploying schema")
	if err := s.Deploy(ctx, &DeployOptions{
		SQL:    opts.Schema,
		Target: db,
	}); err != nil {
		return err
	}

	return nil
}

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
	if err != nil {
		return err
	}
	defer db.Close()

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
