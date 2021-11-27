package cli

import (
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/squire"
)

type DiffCommand struct {
	*baseCommand

	production bool
	verifyDump bool
}

func (c *DiffCommand) Run(args []string) int {
	ctx := c.Ctx
	L := c.Log.Named("diff")

	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Default target URI is empty, which forces Diff to use our dev container.
	var targetURI string

	// If we specified production, then get that.
	if c.production {
		L.Info("diffing against production")
		u, err := c.Config.ProdURL()
		if err != nil {
			return c.exitError(err)
		}

		targetURI = u
	} else {
		L.Info("diffing against development container")
	}

	// Verify our container is running
	err := c.Squire.Diff(ctx, &squire.DiffOptions{
		TargetURI: targetURI,

		// Output verbose info if we have any verbosity set on our logger.
		Verbose: c.Log.IsDebug(),

		Verify: c.verifyDump,
	})
	if err != nil {
		return c.exitError(err)
	}

	return 0
}

func (c *DiffCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		f := sets.NewSet("Command Options")

		f.BoolVar(&flag.BoolVar{
			Name:    "production",
			Target:  &c.production,
			Default: false,
			Usage:   "Diff against the production database.",
			Aliases: []string{"p"},
		})

		f.BoolVar(&flag.BoolVar{
			Name:    "verify-dump",
			Target:  &c.verifyDump,
			Default: false,
			Usage: "Test the diff against a pg_dump clone to verify that " +
				"it produces the expected schema. This isn't 100% accurate and " +
				"so it is disabled by default. Scrutinize any pass/fail results.",
		})
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

  WARNING: The diff is not perfect and does not support all PostgreSQL
  functionality. All common operations are fully supported but there are
  various edges of PostgreSQL that aren't covered. Always manually verify
  diffs.

` + c.Flags().Help())
}
