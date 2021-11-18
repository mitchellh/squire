package cli

import (
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
)

type DownCommand struct {
	*baseCommand

	sqlDir string
	write  bool
}

func (c *DownCommand) Run(args []string) int {
	ctx := c.Ctx

	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Get our container
	ctr, err := c.container()
	if err != nil {
		return c.exitError(err)
	}

	// Launch it
	if err := ctr.Down(ctx); err != nil {
		return c.exitError(err)
	}

	return 0
}

func (c *DownCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		// Nothing today
	})
}

func (c *DownCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *DownCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *DownCommand) Synopsis() string {
	return "Stop and delete PostgreSQL database container"
}

func (c *DownCommand) Help() string {
	return formatHelp(`
Usage: squire down [options]

  Stop and delete the PostgreSQL container started by "up".

  This will remove all data associated with the database. If you are destroying
  only to apply a new schema, try the "squire reset" command instead. If you
  are using a docker-compose.yml file, this will bring down all services in
  the project currently.

` + c.Flags().Help())
}
