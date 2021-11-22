package cli

import (
	"fmt"

	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
)

type URLCommand struct {
	*baseCommand
}

func (c *URLCommand) Run(args []string) int {
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

	fmt.Println(ctr.ConnURI())
	return 0
}

func (c *URLCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		// Nothing today
	})
}

func (c *URLCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *URLCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *URLCommand) Synopsis() string {
	return "Start PostgreSQL database container"
}

func (c *URLCommand) Help() string {
	return formatHelp(`
Usage: squire url [options]

  Outputs the URL to the database that can be used with any PostgreSQL client.

  This outputs the URL to the development database by default. The database
  may or may not be up (you must call "squire up" to bring it up).

  By specifying the "-production" flag, the connection URL to the production
  database will be printed to stdout. This is useful to test that Squire is
  connecting to the proper production database.

` + c.Flags().Help())
}
