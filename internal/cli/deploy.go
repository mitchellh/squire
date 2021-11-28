package cli

import (
	"bytes"
	"database/sql"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cockroachdb/errors"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/dbcontainer"
	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/squire"
)

type DeployCommand struct {
	*baseCommand

	force      bool
	production bool
	sqlPath    string
}

func (c *DeployCommand) Run(args []string) int {
	ctx := c.Ctx
	L := c.Log.Named("deploy")

	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Let's determine our target.
	var targetURI string
	var targetDB *sql.DB

	if c.production {
		L.Warn("deploying to production")
		u, err := c.Config.ProdURL()
		if err != nil {
			return c.exitError(err)
		}

		targetURI = u
	}

	if targetURI == "" {
		L.Debug("no target URI, using dev container by default")
		// Our target is our dev container by default.
		ctr, err := c.Squire.Container()
		if err != nil {
			return c.exitError(err)
		}

		st, err := ctr.Status(ctx)
		if err != nil {
			return c.exitError(err)
		}

		if st.State != dbcontainer.Running {
			return c.exitError(errors.WithDetail(
				errors.New("database container is not running"),
				strings.TrimSpace(errDetailDeployNotRunning),
			))
		}

		targetURI = ctr.ConnURI()
	}

	// Get our SQL reader. Default nil will use the diff.
	var sqlR io.Reader
	if c.sqlPath != "" {
		f, err := os.Open(c.sqlPath)
		if err != nil {
			return c.exitError(err)
		}
		defer f.Close()

		sqlR = f
	}

	// Connect to the database
	L.Debug("target URI", "uri", targetURI)
	targetDB, err := sql.Open("pgx", targetURI)
	if err != nil {
		return c.exitError(err)
	}
	defer targetDB.Close()
	err = backoff.Retry(func() error {
		return targetDB.Ping()
	}, backoff.WithContext(
		backoff.NewConstantBackOff(15*time.Millisecond),
		ctx,
	))
	if err != nil {
		return c.exitError(err)
	}

	// Run our diff
	if sqlR == nil {
		var diff bytes.Buffer
		L.Debug("starting diff")
		err = c.Squire.Diff(ctx, &squire.DiffOptions{
			TargetURI: targetURI,

			// Capture the diff
			Output: &diff,

			// Output verbose info if we have any verbosity set on our logger.
			Verbose: c.Log.IsDebug(),
		})
		if err != nil {
			return c.exitError(err)
		}

		sqlR = &diff
	}

	// Output and verify with user
	if !c.force {
		// TODO
	} else {
		L.Info("force requested, will not ask for user confirmation")
	}

	// Deploy the diff
	L.Debug("starting deploy")
	if err := c.Squire.Deploy(ctx, &squire.DeployOptions{
		SQL:    sqlR,
		Target: targetDB,
	}); err != nil {
		return c.exitError(err)
	}

	colorSuccess.Println("Changes successfully deployed.")
	return 0
}

func (c *DeployCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		f := sets.NewSet("Command Options")

		f.BoolVar(&flag.BoolVar{
			Name:    "force",
			Target:  &c.force,
			Default: false,
			Usage:   "Do not ask for confirmation.",
			Aliases: []string{"f"},
		})

		f.BoolVar(&flag.BoolVar{
			Name:    "production",
			Target:  &c.production,
			Default: false,
			Usage:   "Deploy to the production database.",
			Aliases: []string{"p"},
		})

		f.StringVar(&flag.StringVar{
			Name:    "sql-path",
			Target:  &c.sqlPath,
			Default: "",
			Usage: "SQL file to run for the deployment. If this isn't specified " +
				"the results of diff will be run. This can be used or custom migrations.",
		})
	})
}

func (c *DeployCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *DeployCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *DeployCommand) Synopsis() string {
	return "Deploy schema changes to an existing database"
}

func (c *DeployCommand) Help() string {
	return formatHelp(`
Usage: squire deploy [options]

  Deploy schema changes to a target database.

  This applies the output from "squire diff" to a target database.
  The target database by default is the development container created
  with "squire up". The target database is production if the "-production"
  flag is specified.

  The "-sql-path" flag can be used to specify a custom SQL file to run
  instead of the "diff" output. This can be used in the case where the "diff"
  is incorrect or any other manual changes need to be applied.

  In development, it is typically faster to use "squire reset" to continously
  delete and reapply the full schema, especially if you don't care about
  having a migration path. Deploy can be used to test a final schema change,
  and then to finally deploy it to production.

` + c.Flags().Help())
}

const (
	errDetailDeployNotRunning = `
"squire deploy" was invoked to target the development container, but
the database container is not running. Please run "squire up" to start
the container. If you meant to target production, you must specify the
"-production" flag.
`
)
