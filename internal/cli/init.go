package cli

import (
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
)

type InitCommand struct {
	*baseCommand
}

func (c *InitCommand) Run(args []string) int {
	return 0
}

func (c *InitCommand) Flags() *flag.Sets {
	return nil
}

func (c *InitCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *InitCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *InitCommand) Synopsis() string {
	return "Initialize a new project"
}

func (c *InitCommand) Help() string {
	return formatHelp(`
Usage: squire init [options]

  Initialize a new project for Squire.

` + c.Flags().Help())
}
