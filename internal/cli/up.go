package cli

import (
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
)

type UpCommand struct {
	*baseCommand
}

func (c *UpCommand) Run(args []string) int {
	ctx := c.Ctx

	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Get our container
	ctr, err := c.Squire.Container()
	if err != nil {
		return c.exitError(err)
	}

	// Launch it
	if err := ctr.Up(ctx); err != nil {
		return c.exitError(err)
	}

	return 0
}

func (c *UpCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		// Nothing today
	})
}

func (c *UpCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *UpCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *UpCommand) Synopsis() string {
	return "Start PostgreSQL database container"
}

func (c *UpCommand) Help() string {
	return formatHelp(`
Usage: squire up [options]

  Start a PostgreSQL container for development.

  This starts a Docker container running PostgreSQL that can be used
  for development. Under the hood, this uses Docker Compose, and you may
  utilize an existing Compose configuration file if it exists.

  If you have a "docker-compose.yml" file in this or any parent
  directories, then Squire will attempt to find a database service by
  looking for a service with the "x-squire" configuration set. If no
  database service is found, Squire will spin up a default PostgreSQL
  container.

  You can destroy the running development database using "squire down".
  If you want to just reapply the schema, you can run "squire reset".

` + c.Flags().Help())
}
