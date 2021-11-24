package cli

import (
	"fmt"

	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
)

type URLCommand struct {
	*baseCommand

	production bool
}

func (c *URLCommand) Run(args []string) int {
	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// If production, grab that
	if c.production {
		uri, err := c.Config.ProdURL()
		if err != nil {
			return c.exitError(err)
		}

		fmt.Println(uri)
		return 0
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
		f := sets.NewSet("Command Options")

		f.BoolVar(&flag.BoolVar{
			Name:    "production",
			Target:  &c.production,
			Default: false,
			Usage:   "Use the production database.",
			Aliases: []string{"p"},
		})
	})
}

func (c *URLCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *URLCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *URLCommand) Synopsis() string {
	return "Output connection URL to the database"
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
