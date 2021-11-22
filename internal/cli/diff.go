package cli

import (
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/squire"
)

type DiffCommand struct {
	*baseCommand
}

func (c *DiffCommand) Run(args []string) int {
	ctx := c.Ctx

	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Verify our container is running
	err := c.Squire.Diff(ctx, &squire.DiffOptions{
		// Output verbose info if we have any verbosity set on our logger.
		Verbose: c.Log.IsDebug(),
	})
	if err != nil {
		return c.exitError(err)
	}

	return 0
}

func (c *DiffCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		// Nothing today
	})
}

func (c *DiffCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *DiffCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *DiffCommand) Synopsis() string {
	return "Apply a fresh schema to the dev database"
}

func (c *DiffCommand) Help() string {
	return formatHelp(`
Usage: squire diff [options]

  Show a SQL diff from the current schema to a deployed schema.

  The diff is a set of SQL statements that should be executed to go
  from a deployed schema to the schema described in the SQL files.
  This can either be applied directly (with "squire deploy") or used
  as a starting point for migration tooling.

  If there is no diff, the output will be empty with an exit code of 0.

  By default, this will show the diff between the current SQL files and
  the deployed schema in the development container from "squire up". In this
  case, the development container must be up and running.

` + c.Flags().Help())
}
