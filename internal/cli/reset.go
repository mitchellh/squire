package cli

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/dbcontainer"
	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/squire"
)

type ResetCommand struct {
	*baseCommand
}

func (c *ResetCommand) Run(args []string) int {
	ctx := c.Ctx

	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Verify our container is running
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
			strings.TrimSpace(errDetailNotRunning),
		))
	}

	// Reset
	if err := c.Squire.Reset(ctx, &squire.ResetOptions{}); err != nil {
		return c.exitError(err)
	}

	colorSuccess.Println("Database was successfully reset.")
	return 0
}

func (c *ResetCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		// Nothing today
	})
}

func (c *ResetCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *ResetCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *ResetCommand) Synopsis() string {
	return "Apply a fresh schema to the dev database"
}

func (c *ResetCommand) Help() string {
	return formatHelp(`
Usage: squire reset [options]

  Reset the database to a fresh schema (destroys all data).

  This deletes the full database, recreates it, and applies the schema.
  This can be used to quickly iterate and test on the schema locally in
  dev. A common workflow is: "squire up", edit SQL files, "squire reset"
  in development (not production). In development you're usually less worried
  about clean diff-based deploys so reset is more appropriate.

  Reset currently only works against the development database. It is not
  possible to reset the production database. You must do this manually
  without the help of Squire; it is too dangerous of an operation to
  make so easily accessible in Squire.

` + c.Flags().Help())
}

const (
	errDetailNotRunning = `
The database container isn't running. Please run "squire up" to start
the container.
`
)
