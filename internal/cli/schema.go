package cli

import (
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
)

type SchemaCommand struct {
	*baseCommand
}

func (c *SchemaCommand) Run(args []string) int {
	return 0
}

func (c *SchemaCommand) Flags() *flag.Sets {
	return nil
}

func (c *SchemaCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *SchemaCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *SchemaCommand) Synopsis() string {
	return "Build SQL schema"
}

func (c *SchemaCommand) Help() string {
	return formatHelp(`
Usage: squire schema [options]

  Output the full SQL schema.

  This builds the SQL schema from the .sql files in your "sql/" directory
  and outputs it to stdout. This is NOT reading the currently deployed
  schema from any database.

` + c.Flags().Help())
}
